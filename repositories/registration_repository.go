package repositories

import (
	"context"
	"errors"
	"strings"

	"event-registration-api/models"
	"event-registration-api/services"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegistrationRepository interface {
	RegisterForEvent(ctx context.Context, eventID, userID uuid.UUID) (*models.Registration, error)
	CancelRegistration(ctx context.Context, registrationID uuid.UUID) error
}

type registrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &registrationRepository{db: db}
}

// RegisterForEvent performs a safe booking using row-level locking setup inside a GORM transaction.
func (r *registrationRepository) RegisterForEvent(ctx context.Context, eventID, userID uuid.UUID) (*models.Registration, error) {
	var registration models.Registration

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var event models.Event

		// 1. SELECT FOR UPDATE is needed because it acquires a row-level lock on the event row.
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&event, "id = ?", eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return services.ErrEventNotFound
			}
			return err
		}

		// 2. Check seats
		if event.AvailableSeats <= 0 {
			return services.ErrNoSeatsAvailable
		}

		// 3. Create registration
		registration = models.Registration{
			EventID: eventID,
			UserID:  userID,
			Status:  "confirmed",
		}

		if err := tx.Create(&registration).Error; err != nil {
			if isUniqueViolation(err) {
				return services.ErrAlreadyRegistered
			}
			return err
		}

		// 4. Update event available seats
		if err := tx.Model(&event).Update("available_seats", gorm.Expr("available_seats - 1")).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &registration, nil
}

func (r *registrationRepository) CancelRegistration(ctx context.Context, registrationID uuid.UUID) error {
	// Cancel registration and increment available seats
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reg models.Registration
		if err := tx.First(&reg, "id = ?", registrationID).Error; err != nil {
			return err
		}

		if reg.Status == "cancelled" {
			return nil
		}

		// acquire lock on event
		var event models.Event
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&event, "id = ?", reg.EventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return services.ErrEventNotFound
			}
			return err
		}

		reg.Status = "cancelled"
		if err := tx.Save(&reg).Error; err != nil {
			return err
		}

		if err := tx.Model(&event).Update("available_seats", gorm.Expr("available_seats + 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

// isUniqueViolation checks whether the error returned by the database
// is a unique constraint violation.
func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key")
}

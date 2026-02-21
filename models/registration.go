package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Registration represents a user's booking for an event.
// The composite unique index on (event_id, user_id) acts as a second-line
// defence against duplicate confirmed bookings. For re-registration after
// cancellation, the repository checks for existing confirmed records
// rather than relying solely on this constraint.
type Registration struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_user" json:"event_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_user" json:"user_id"`
	Status    string    `gorm:"not null;default:'confirmed'" json:"status"` // 'confirmed' or 'cancelled'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate generates a UUID in Go before inserting.
func (r *Registration) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

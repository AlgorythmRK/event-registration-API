package repositories

import (
	"context"

	"event-registration-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	GetEvents(ctx context.Context) ([]models.Event, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *eventRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) GetEvents(ctx context.Context) ([]models.Event, error) {
	var events []models.Event
	err := r.db.WithContext(ctx).Find(&events).Error
	return events, err
}

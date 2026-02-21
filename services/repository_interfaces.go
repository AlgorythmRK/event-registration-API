package services

import (
	"context"

	"event-registration-api/models"

	"github.com/google/uuid"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	GetEvents(ctx context.Context) ([]models.Event, error)
}

type RegistrationRepository interface {
	RegisterForEvent(ctx context.Context, eventID, userID uuid.UUID) (*models.Registration, error)
	CancelRegistration(ctx context.Context, registrationID uuid.UUID) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUsers(ctx context.Context) ([]models.User, error)
}

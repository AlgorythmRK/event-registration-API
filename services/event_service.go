package services

import (
	"context"

	"event-registration-api/models"

	"github.com/google/uuid"
)

type EventService interface {
	CreateEvent(ctx context.Context, organizerID uuid.UUID, event *models.Event) error
	GetEvents(ctx context.Context) ([]models.Event, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
}

type eventService struct {
	repo     EventRepository
	userRepo UserRepository
}

func NewEventService(repo EventRepository, userRepo UserRepository) EventService {
	return &eventService{repo: repo, userRepo: userRepo}
}

func (s *eventService) CreateEvent(ctx context.Context, organizerID uuid.UUID, event *models.Event) error {
	user, err := s.userRepo.GetUserByID(ctx, organizerID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Role != "organizer" {
		return ErrNotOrganizer
	}

	event.OrganizerID = organizerID
	return s.repo.CreateEvent(ctx, event)
}

func (s *eventService) GetEvents(ctx context.Context) ([]models.Event, error) {
	return s.repo.GetEvents(ctx)
}

func (s *eventService) GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}

package services

import (
	"context"

	"event-registration-api/models"

	"github.com/google/uuid"
)

type RegistrationService interface {
	RegisterForEvent(ctx context.Context, eventID, userID uuid.UUID) (*models.Registration, error)
	CancelRegistration(ctx context.Context, registrationID uuid.UUID) error
}

type registrationService struct {
	repo      RegistrationRepository
	eventRepo EventRepository
	userRepo  UserRepository
}

func NewRegistrationService(
	repo RegistrationRepository,
	eventRepo EventRepository,
	userRepo UserRepository,
) RegistrationService {
	return &registrationService{
		repo:      repo,
		eventRepo: eventRepo,
		userRepo:  userRepo,
	}
}

func (s *registrationService) RegisterForEvent(ctx context.Context, eventID, userID uuid.UUID) (*models.Registration, error) {
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	_, err = s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}

	return s.repo.RegisterForEvent(ctx, eventID, userID)
}

func (s *registrationService) CancelRegistration(ctx context.Context, registrationID uuid.UUID) error {
	return s.repo.CancelRegistration(ctx, registrationID)
}

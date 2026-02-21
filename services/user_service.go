package services

import (
	"context"

	"event-registration-api/models"

	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	// Let repository layer generate UUID and handle defaults
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) GetUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.GetUsers(ctx)
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		// Example of mapping to sentinel
		return nil, ErrUserNotFound
	}
	return user, nil
}

package services

import (
	"errors"
	"todo-backend/internal/models"
	"todo-backend/internal/repositories"

	"github.com/google/uuid"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

// NewUserService creates a new UserService
func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

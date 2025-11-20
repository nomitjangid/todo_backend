package services

import (
	"errors"
	"testing"
	"time"
	"todo-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}


func TestAuthService_RegisterUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	authService := NewAuthService(mockUserRepo) // Inject mock

	t.Run("successfully registers a user", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"

		mockUserRepo.On("GetUserByEmail", email).Return(nil, errors.New("not found")).Once()
		mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil).Once()

		user, err := authService.RegisterUser(email, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.NotEmpty(t, user.PasswordHash)
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		assert.NoError(t, err) // Verify password was hashed correctly

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error if user already exists", func(t *testing.T) {
		email := "existing@example.com"
		password := "password123"
		existingUser := &models.User{ID: uuid.New(), Email: email, PasswordHash: "hashed", CreatedAt: time.Now()}

		mockUserRepo.On("GetUserByEmail", email).Return(existingUser, nil).Once()

		user, err := authService.RegisterUser(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.EqualError(t, err, "user already exists")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error if CreateUser fails", func(t *testing.T) {
		email := "fail@example.com"
		password := "password123"

		mockUserRepo.On("GetUserByEmail", email).Return(nil, errors.New("not found")).Once()
		mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(errors.New("db error")).Once()

		user, err := authService.RegisterUser(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.EqualError(t, err, "db error")

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_LoginUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	authService := NewAuthService(mockUserRepo) // Inject mock

	// Hash a password for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:           uuid.New(),
		Email:        "login@example.com",
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	t.Run("successfully logs in a user", func(t *testing.T) {
		email := "login@example.com"
		password := "password123"

		mockUserRepo.On("GetUserByEmail", email).Return(testUser, nil).Once()

		token, err := authService.LoginUser(email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error for invalid email", func(t *testing.T) {
		email := "nonexistent@example.com"
		password := "password123"

		mockUserRepo.On("GetUserByEmail", email).Return(nil, errors.New("not found")).Once()

		token, err := authService.LoginUser(email, password)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.EqualError(t, err, "invalid credentials")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error for invalid password", func(t *testing.T) {
		email := "login@example.com"
		password := "wrongpassword"

		mockUserRepo.On("GetUserByEmail", email).Return(testUser, nil).Once()

		token, err := authService.LoginUser(email, password)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.EqualError(t, err, "invalid credentials")

		mockUserRepo.AssertExpectations(t)
	})
}

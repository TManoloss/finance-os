package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository é um mock para a interface UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePluggyCredentials(ctx context.Context, userID, clientID, clientSecret string) error {
	args := m.Called(ctx, userID, clientID, clientSecret)
	return args.Error(0)
}

func (m *MockUserRepository) GetPluggyCredentials(ctx context.Context, userID string) (string, string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func setupTest() (*AuthService, *MockUserRepository) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWTSecret: "test_secret",
	}
	jwtSvc := NewJWTService(cfg)
	authSvc := NewAuthService(mockRepo, jwtSvc)
	return authSvc, mockRepo
}

func TestAuthService_Register(t *testing.T) {
	t.Run("sucesso no registro", func(t *testing.T) {
		svc, mockRepo := setupTest()
		ctx := context.Background()

		mockRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)
		mockRepo.On("SaveRefreshToken", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		res, err := svc.Register(ctx, "Test User", "new@example.com", "Senha@123456")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Test User", res.User.Name)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("falha por email já existente", func(t *testing.T) {
		svc, mockRepo := setupTest()
		ctx := context.Background()

		mockRepo.On("FindByEmail", ctx, "existing@example.com").Return(&models.User{Email: "existing@example.com"}, nil)

		res, err := svc.Register(ctx, "Test User", "existing@example.com", "Senha@123456")

		assert.Error(t, err)
		assert.Equal(t, ErrEmailTaken, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("falha por senha fraca", func(t *testing.T) {
		svc, _ := setupTest()
		ctx := context.Background()

		res, err := svc.Register(ctx, "Test User", "new@example.com", "123")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPassword, err)
		assert.Nil(t, res)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("sucesso no login", func(t *testing.T) {
		svc, mockRepo := setupTest()
		ctx := context.Background()

		// "Senha@123456" hashed with cost 12
		password := "Senha@123456"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

		user := &models.User{
			ID:           "user-123",
			Email:        "user@example.com",
			PasswordHash: string(hashedPassword),
		}

		mockRepo.On("FindByEmail", ctx, "user@example.com").Return(user, nil)
		mockRepo.On("SaveRefreshToken", ctx, "user-123", mock.Anything, mock.Anything).Return(nil)

		res, err := svc.Login(ctx, "user@example.com", password)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, user.ID, res.User.ID)
		assert.NotEmpty(t, res.AccessToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("falha por credenciais inválidas", func(t *testing.T) {
		svc, mockRepo := setupTest()
		ctx := context.Background()

		mockRepo.On("FindByEmail", ctx, "user@example.com").Return(nil, errors.New("not found"))

		res, err := svc.Login(ctx, "user@example.com", "wrong-password")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Nil(t, res)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("sucesso no refresh", func(t *testing.T) {
		svc, mockRepo := setupTest()
		ctx := context.Background()

		oldToken := "old-refresh-token"
		userID := "user-123"
		user := &models.User{ID: userID, Email: "user@example.com"}

		mockRepo.On("ValidateRefreshToken", ctx, oldToken).Return(userID, nil)
		mockRepo.On("FindByID", ctx, userID).Return(user, nil)
		mockRepo.On("DeleteRefreshToken", ctx, oldToken).Return(nil)
		mockRepo.On("SaveRefreshToken", ctx, userID, mock.Anything, mock.Anything).Return(nil)

		res, err := svc.RefreshToken(ctx, oldToken)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.NotEqual(t, oldToken, res.RefreshToken)
		mockRepo.AssertExpectations(t)
	})
}

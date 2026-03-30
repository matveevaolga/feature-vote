package service

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupUserTest(t *testing.T) (*mocks.MockUserRepository, *mocks.MockJWTService, *UserService) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWTService := new(mocks.MockJWTService)
	service := NewUserService(mockUserRepo, mockJWTService)
	return mockUserRepo, mockJWTService, service
}

func TestUserService_CreateUser_Success(t *testing.T) {
	mockUserRepo, _, service := setupUserTest(t)

	username := "testuser"
	email := "testuser@temp.local"
	userID := uuid.Must(uuid.NewV4())

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, domain.ErrUserNotFound)
	mockUserRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, domain.ErrUserNotFound)
	mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	user := &domain.User{
		ID:       userID,
		Username: username,
		Email:    email,
	}
	err := service.CreateUser(context.Background(), user)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_AlreadyExists(t *testing.T) {
	mockUserRepo, _, service := setupUserTest(t)

	username := "existinguser"
	email := "existinguser@temp.local"
	existingUser := &domain.User{Username: username, Email: email}

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(existingUser, nil)

	user := &domain.User{
		ID:       uuid.Must(uuid.NewV4()),
		Username: username,
		Email:    email,
	}
	err := service.CreateUser(context.Background(), user)

	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	mockUserRepo, _, service := setupUserTest(t)

	userID := uuid.Must(uuid.NewV4())
	expectedUser := &domain.User{ID: userID, Username: "testuser", Email: "test@example.com"}

	mockUserRepo.On("GetUserByID", mock.Anything, userID.String()).Return(expectedUser, nil)

	user, err := service.GetUserByID(context.Background(), userID.String())

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockUserRepo, _, service := setupUserTest(t)

	userID := uuid.Must(uuid.NewV4())

	mockUserRepo.On("GetUserByID", mock.Anything, userID.String()).Return(nil, domain.ErrUserNotFound)

	user, err := service.GetUserByID(context.Background(), userID.String())

	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Register_Success(t *testing.T) {
	mockUserRepo, _, service := setupUserTest(t)

	username := "newuser"
	email := "newuser@example.com"
	password := "password123"

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, domain.ErrUserNotFound)
	mockUserRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, domain.ErrUserNotFound)
	mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := service.Register(context.Background(), RegisterParams{
		Username: username,
		Email:    email,
		Password: password,
	})

	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	mockUserRepo, mockJWT, service := setupUserTest(t)

	userID := uuid.Must(uuid.NewV4())
	email := "user@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:           userID,
		Username:     "testuser",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)
	mockJWT.On("GenerateToken", userID).Return("fake-jwt-token", nil)

	_, token, err := service.Login(context.Background(), LoginParams{
		Email:    email,
		Password: password,
	})

	assert.NoError(t, err)
	assert.Equal(t, "fake-jwt-token", token)
	mockUserRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	mockUserRepo, _, service := setupUserTest(t)

	email := "user@example.com"
	password := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	user := &domain.User{
		ID:           uuid.Must(uuid.NewV4()),
		Username:     "testuser",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)

	_, _, err := service.Login(context.Background(), LoginParams{
		Email:    email,
		Password: password,
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

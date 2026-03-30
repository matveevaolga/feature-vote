package service

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	username := "testuser"
	userID := uuid.Must(uuid.NewV4())

	mockRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, domain.ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	err := service.CreateUser(context.Background(), &domain.User{
		ID:       userID,
		Username: username,
	})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_AlreadyExists(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	username := "existinguser"
	existingUser := &domain.User{Username: username}

	mockRepo.On("GetUserByUsername", mock.Anything, username).Return(existingUser, nil)

	err := service.CreateUser(context.Background(), &domain.User{
		ID:       uuid.Must(uuid.NewV4()),
		Username: username,
	})

	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.Must(uuid.NewV4())
	expectedUser := &domain.User{ID: userID, Username: "testuser"}

	mockRepo.On("GetUserByID", mock.Anything, userID.String()).Return(expectedUser, nil)

	user, err := service.GetUserByID(context.Background(), userID.String())

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.Must(uuid.NewV4())

	mockRepo.On("GetUserByID", mock.Anything, userID.String()).Return(nil, domain.ErrUserNotFound)

	user, err := service.GetUserByID(context.Background(), userID.String())

	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

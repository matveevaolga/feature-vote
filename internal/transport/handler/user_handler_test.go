package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_CreateUser_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockUserRepo)
	handler := NewUserHandler(userService)

	username := "newuser"

	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, domain.ErrUserNotFound)
	mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	reqBody := dto.CreateUserRequest{Username: username}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateUser)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response dto.UserResponse
	ParseJSONResponse(t, rr, &response)
	assert.Equal(t, username, response.Username)
	assert.NotEmpty(t, response.ID)

	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockUserRepo)
	handler := NewUserHandler(userService)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateUser)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_CreateUser_EmptyUsername(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockUserRepo)
	handler := NewUserHandler(userService)

	reqBody := dto.CreateUserRequest{Username: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateUser)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_CreateUser_UsernameTooShort(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockUserRepo)
	handler := NewUserHandler(userService)

	reqBody := dto.CreateUserRequest{Username: "ab"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateUser)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_CreateUser_AlreadyExists(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockUserRepo)
	handler := NewUserHandler(userService)

	username := "existinguser"

	existingUser := &domain.User{Username: username}
	mockUserRepo.On("GetUserByUsername", mock.Anything, username).Return(existingUser, nil)

	reqBody := dto.CreateUserRequest{Username: username}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateUser)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockUserRepo.AssertNotCalled(t, "CreateUser")
}

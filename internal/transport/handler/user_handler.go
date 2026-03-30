package handler

import (
	"encoding/json"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	uuid "github.com/gofrs/uuid/v5"

	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
)

type UserHandler struct {
	userService *service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: invalid username format", err)
		return
	}
	uuid4, err := uuid.NewV4()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate user ID", err)
		return
	}

	user := &domain.User{
		ID:        uuid4,
		Username:  req.Username,
		CreatedAt: time.Now(),
	}
	if err := h.userService.CreateUser(r.Context(), user); err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			RespondWithError(w, http.StatusConflict, "User with this username already exists", err)
		case domain.ErrInvalidUsername:
			RespondWithError(w, http.StatusBadRequest, "Invalid username format", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
		}
		return
	}

	response := dto.UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
	RespondWithJSON(w, http.StatusCreated, response)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}

	user, err := h.userService.Register(r.Context(), service.RegisterParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			RespondWithError(w, http.StatusConflict, "User already exists", err)
		case domain.ErrInvalidUsername, domain.ErrInvalidEmail, domain.ErrInvalidPassword:
			RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to register user", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, dto.UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}

	user, token, err := h.userService.Login(r.Context(), service.LoginParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to login", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.LoginResponse{
		User: dto.UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	})
}

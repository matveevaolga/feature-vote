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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: invalid username format", http.StatusBadRequest)
		return
	}
	uuid4, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
			http.Error(w, "User with this username already exists", http.StatusConflict)
		case domain.ErrInvalidUsername:
			http.Error(w, "Invalid username format", http.StatusBadRequest)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	response := dto.UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

type GroupHandler struct {
	groupService *service.GroupService
	userService  *service.UserService
	validate     *validator.Validate
}

func NewGroupHandler(groupService *service.GroupService, userService *service.UserService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		userService:  userService,
		validate:     validator.New(),
	}
}

// handles POST /groups
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userIDString := r.Context().Value(middleware.UserIDKey).(string)
	ownerID, err := uuid.FromString(userIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	var req dto.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}

	group, err := h.groupService.CreateGroup(r.Context(), req.Name, ownerID)
	if err != nil {
		switch err {
		case domain.ErrInvalidGroupName:
			RespondWithError(w, http.StatusBadRequest, "Invalid group name", err)
		case domain.ErrUserNotFound:
			RespondWithError(w, http.StatusNotFound, "User not found", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create group", err)
		}
		return
	}
	RespondWithJSON(w, http.StatusCreated, dto.GroupResponse{
		ID:        group.ID.String(),
		Name:      group.Name,
		OwnerID:   group.OwnerID.String(),
		CreatedAt: group.CreatedAt,
	})
}

// handles GET /groups/{id}
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	group, err := h.groupService.GetGroupByID(r.Context(), groupID)
	if err != nil {
		switch err {
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get group", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.GroupResponse{
		ID:        group.ID.String(),
		Name:      group.Name,
		OwnerID:   group.OwnerID.String(),
		CreatedAt: group.CreatedAt,
	})
}

// handles PUT /groups/{id}
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	var req dto.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}

	err = h.groupService.UpdateGroup(r.Context(), groupID, req.Name, userID)
	if err != nil {
		switch err {
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		case domain.ErrNotGroupOwner:
			RespondWithError(w, http.StatusForbidden, "Only owner can update group", err)
		case domain.ErrInvalidGroupName:
			RespondWithError(w, http.StatusBadRequest, "Invalid group name", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to update group", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handles DELETE /groups/{id}
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}

	err = h.groupService.DeleteGroup(r.Context(), groupID, userID)
	if err != nil {
		switch err {
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		case domain.ErrNotGroupOwner:
			RespondWithError(w, http.StatusForbidden, "Only owner can delete group", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete group", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handles GET /users/groups
func (h *GroupHandler) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	groups, err := h.groupService.GetUserGroups(r.Context(), userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user groups", err)
		return
	}
	response := make([]dto.GroupResponse, len(groups))
	for i, group := range groups {
		response[i] = dto.GroupResponse{
			ID:        group.ID.String(),
			Name:      group.Name,
			OwnerID:   group.OwnerID.String(),
			CreatedAt: group.CreatedAt,
		}
	}
	RespondWithJSON(w, http.StatusOK, response)
}

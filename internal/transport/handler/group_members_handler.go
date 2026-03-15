package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

// handles POST /groups/{id}/invite
func (h *GroupHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	ownerIDString := r.Context().Value(middleware.UserIDKey).(string)
	ownerID, err := uuid.FromString(ownerIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	var req dto.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}
	userID, err := uuid.FromString(req.UserID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}
	err = h.groupService.InviteMember(r.Context(), groupID, userID, ownerID)
	if err != nil {
		switch err {
		case domain.ErrNotGroupOwner, domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusForbidden, "Not enough permissions", err)
		case domain.ErrUserNotFound:
			RespondWithError(w, http.StatusNotFound, "User not found", err)
		case domain.ErrAlreadyGroupMember:
			RespondWithError(w, http.StatusConflict, "User is already a member", err)
		case domain.ErrInvitationAlreadySent:
			RespondWithError(w, http.StatusConflict, "Invitation already sent", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to invite member", err)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// handles GET /groups/{id}/members
func (h *GroupHandler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
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
	members, err := h.groupService.GetGroupMembers(r.Context(), groupID, userID)
	if err != nil {
		switch err {
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusForbidden, "Not a group member", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get members", err)
		}
		return
	}
	response := make([]dto.GroupMemberResponse, len(members))
	for i, member := range members {
		user, err := h.userService.GetUserByID(r.Context(), member.UserID.String())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get members", err)
			return
		}
		response[i] = dto.GroupMemberResponse{
			UserID:   member.UserID.String(),
			Username: user.Username,
			Role:     string(member.Role),
			JoinedAt: member.JoinedAt,
		}
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// handles DELETE /groups/{id}/members/{userID}
func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ownerIDString := r.Context().Value(middleware.UserIDKey).(string)
	ownerID, err := uuid.FromString(ownerIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	memberIDStr := chi.URLParam(r, "userID")
	if memberIDStr == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid member ID", nil)
		return
	}
	memberID, err := uuid.FromString(memberIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid member ID", err)
		return
	}
	err = h.groupService.RemoveMember(r.Context(), groupID, ownerID, memberID)
	if err != nil {
		switch err {
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		case domain.ErrNotGroupOwner:
			RespondWithError(w, http.StatusForbidden, "Only owner can remove members", err)
		case domain.ErrCannotRemoveGroupOwner:
			RespondWithError(w, http.StatusBadRequest, "Cannot remove group owner", err)
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusNotFound, "Member not found in group", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to remove member", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handles POST /groups/{id}/leave
func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
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

	err = h.groupService.LeaveGroup(r.Context(), groupID, userID)
	if err != nil {
		switch err {
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusNotFound, "Not a group member", err)
		case domain.ErrCannotRemoveGroupOwner:
			RespondWithError(w, http.StatusBadRequest, "Owner cannot leave, transfer ownership first", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to leave group", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handles PUT /groups/{id}/members/{userID}/role
func (h *GroupHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	ownerIDString := r.Context().Value(middleware.UserIDKey).(string)
	ownerID, err := uuid.FromString(ownerIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	memberIDStr := chi.URLParam(r, "userID")
	if memberIDStr == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid member ID", nil)
		return
	}
	memberID, err := uuid.FromString(memberIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid member ID", err)
		return
	}
	var req struct {
		Role string `json:"role" validate:"required,oneof=admin member"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}
	var newRole domain.Role
	switch req.Role {
	case "admin":
		newRole = domain.RoleAdmin
	case "member":
		newRole = domain.RoleMember
	default:
		RespondWithError(w, http.StatusBadRequest, "Invalid role", nil)
		return
	}

	err = h.groupService.UpdateMemberRole(r.Context(), groupID, ownerID, memberID, newRole)
	if err != nil {
		switch err {
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		case domain.ErrNotGroupOwner:
			RespondWithError(w, http.StatusForbidden, "Only owner can update roles", err)
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusNotFound, "Member not found in group", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to update role", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handles POST /groups/{id}/transfer
func (h *GroupHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	ownerIDString := r.Context().Value(middleware.UserIDKey).(string)
	ownerID, err := uuid.FromString(ownerIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	var req dto.TransferOwnershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	newOwnerID, err := uuid.FromString(req.NewOwnerID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid new owner ID", err)
		return
	}
	err = h.groupService.TransferOwnership(r.Context(), groupID, ownerID, newOwnerID)
	if err != nil {
		switch err {
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		case domain.ErrNotGroupOwner:
			RespondWithError(w, http.StatusForbidden, "Only owner can transfer ownership", err)
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusNotFound, "New owner must be a group member", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to transfer ownership", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

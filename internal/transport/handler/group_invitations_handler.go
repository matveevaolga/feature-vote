package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

// handles GET /users/invitations
func (h *GroupHandler) GetInvitations(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	_, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	invitations, err := h.groupService.GetUserInvitations(r.Context(), userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get invitations", err)
		return
	}
	response := make([]dto.InvitationResponse, len(invitations))
	for i, inv := range invitations {
		response[i] = dto.InvitationResponse{
			ID:        inv.ID.String(),
			GroupID:   inv.GroupID.String(),
			GroupName: inv.GroupName,
			UserID:    inv.UserID.String(),
			Status:    string(inv.Status),
			CreatedAt: inv.CreatedAt,
			UpdatedAt: inv.UpdatedAt,
		}
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// handles POST /invitations/{id}/accept
func (h *GroupHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	invitationID := chi.URLParam(r, "id")
	if invitationID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid invitation ID", nil)
		return
	}

	err = h.groupService.AcceptInvitation(r.Context(), invitationID, userID)
	if err != nil {
		switch err {
		case domain.ErrInvitationNotFound:
			RespondWithError(w, http.StatusNotFound, "Invitation not found", err)
		case domain.ErrInvitationNotPending:
			RespondWithError(w, http.StatusConflict, "Invitation is not pending", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to accept invitation", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handles POST /invitations/{id}/decline
func (h *GroupHandler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	invitationID := chi.URLParam(r, "id")
	if invitationID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid invitation ID", nil)
		return
	}

	err = h.groupService.DeclineInvitation(r.Context(), invitationID, userID)
	if err != nil {
		switch err {
		case domain.ErrInvitationNotFound:
			RespondWithError(w, http.StatusNotFound, "Invitation not found", err)
		case domain.ErrInvitationNotPending:
			RespondWithError(w, http.StatusConflict, "Invitation is not pending", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to decline invitation", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

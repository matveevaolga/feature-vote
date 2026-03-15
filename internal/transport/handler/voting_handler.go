package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

type VotingHandler struct {
	votingService *service.VotingService
	validate      *validator.Validate
}

func NewVotingHandler(votingService *service.VotingService) *VotingHandler {
	return &VotingHandler{
		votingService: votingService,
		validate:      validator.New(),
	}
}

// handles POST /votings
func (h *VotingHandler) CreateVoting(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	var req dto.CreateVotingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}
	voting, err := h.votingService.CreateVoting(
		r.Context(),
		req.GroupID,
		userID,
		req.FeatureName,
		req.Description,
		time.Duration(req.DurationSec)*time.Second,
	)
	if err != nil {
		switch err {
		case domain.ErrNotGroupOwner, domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusForbidden, "Not enough permissions", err)
		case domain.ErrGroupNotFound:
			RespondWithError(w, http.StatusNotFound, "Group not found", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create voting", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, dto.CreateVotingResponse{
		ID:          voting.ID.String(),
		GroupID:     voting.GroupID.String(),
		FeatureName: voting.FeatureName,
		Description: voting.Description,
		Status:      string(voting.Status),
		StartedAt:   voting.StartedAt,
		EndsAt:      voting.EndsAt,
	})
}

// handles POST /votings/{id}/votes
func (h *VotingHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	votingID := chi.URLParam(r, "id")
	if votingID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid voting ID", nil)
		return
	}
	var req dto.CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Validation failed: "+err.Error(), err)
		return
	}
	var voteType domain.VoteType
	switch req.Vote {
	case "yes":
		voteType = domain.VoteYes
	case "no":
		voteType = domain.VoteNo
	default:
		RespondWithError(w, http.StatusBadRequest, "Invalid vote type", nil)
		return
	}

	err = h.votingService.CastVote(r.Context(), votingID, userID, voteType)
	if err != nil {
		switch err {
		case domain.ErrVotingNotFound:
			RespondWithError(w, http.StatusNotFound, "Voting not found", err)
		case domain.ErrVotingNotActive:
			RespondWithError(w, http.StatusConflict, "Voting is not active", err)
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusForbidden, "You are not a member of this group", err)
		case domain.ErrAlreadyVoted:
			RespondWithError(w, http.StatusConflict, "You have already voted", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to cast vote", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "vote recorded"})
}

// handles POST /votings/{id}/stop
func (h *VotingHandler) StopVoting(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	votingID := chi.URLParam(r, "id")
	if votingID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid voting ID", nil)
		return
	}

	err = h.votingService.StopVoting(r.Context(), votingID, userID)
	if err != nil {
		switch err {
		case domain.ErrVotingNotFound:
			RespondWithError(w, http.StatusNotFound, "Voting not found", err)
		case domain.ErrNotGroupOwner:
			RespondWithError(w, http.StatusForbidden, "Only owner or admin can stop voting", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to stop voting", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "voting stopped"})
}

// handles GET /votings/{id}/results
func (h *VotingHandler) GetVotingResults(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	votingID := chi.URLParam(r, "id")
	if votingID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid voting ID", nil)
		return
	}
	result, err := h.votingService.GetVotingResult(r.Context(), votingID, userID)
	if err != nil {
		switch err {
		case domain.ErrVotingNotFound:
			RespondWithError(w, http.StatusNotFound, "Voting not found", err)
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusForbidden, "You are not a member of this group", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get results", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.VotingResultResponse{
		VotingID:     result.VotingID.String(),
		TotalVotes:   result.TotalVotes,
		YesVotes:     result.YesVotes,
		NoVotes:      result.NoVotes,
		YesPercent:   result.YesPercent,
		NoPercent:    result.NoPercent,
		Participated: result.Participated,
		TotalMembers: result.TotalMembers,
		IsCompleted:  result.IsCompleted,
	})
}

// handles GET /votings/{id}
func (h *VotingHandler) GetVotingStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	votingID := chi.URLParam(r, "id")
	if votingID == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid voting ID", nil)
		return
	}
	voting, err := h.votingService.GetVotingByID(r.Context(), votingID, userID)
	if err != nil {
		switch err {
		case domain.ErrVotingNotFound:
			RespondWithError(w, http.StatusNotFound, "Voting not found", err)
		case domain.ErrNotGroupMember:
			RespondWithError(w, http.StatusForbidden, "You are not a member of this group", err)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to get voting status", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.VotingStatusResponse{
		ID:          voting.ID.String(),
		GroupID:     voting.GroupID.String(),
		FeatureName: voting.FeatureName,
		Description: voting.Description,
		Status:      string(voting.Status),
		StartedAt:   voting.StartedAt,
		EndsAt:      voting.EndsAt,
		CompletedAt: voting.CompletedAt,
		Result:      voting.Result,
	})
}

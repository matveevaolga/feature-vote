package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupVotingTest(t *testing.T) (*service.VotingService, *mocks.MockVotingRepository, *mocks.MockGroupRepository, *VotingHandler) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	votingService := service.NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)
	handler := NewVotingHandler(votingService)

	return votingService, mockVotingRepo, mockGroupRepo, handler
}

func TestVotingHandler_CastVote_Success(t *testing.T) {
	votingService, _, mockGroupRepo, handler := setupVotingTest(t)

	userID := uuid.Must(uuid.NewV4())
	votingID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())
	createdBy := uuid.Must(uuid.NewV4())

	voting := &domain.Voting{
		ID:          votingID,
		GroupID:     groupID,
		CreatedBy:   createdBy,
		FeatureName: "Test Feature",
		Description: "Test Description",
		Status:      domain.VotingStatusActive,
		Duration:    5 * time.Minute,
		StartedAt:   time.Now(),
		EndsAt:      time.Now().Add(5 * time.Minute),
	}

	voteChan := make(chan *domain.Vote, 100)
	activeVoting := &service.ActiveVoting{
		Voting:   voting,
		VoteChan: voteChan,
		StopChan: make(chan struct{}),
		Votes:    make(map[uuid.UUID]domain.VoteType),
		Ctx:      context.Background(),
		Cancel:   func() {},
	}

	votingService.AddActiveVotingForTest(votingID.String(), activeVoting)

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)

	reqBody := dto.CastVoteRequest{Vote: "yes"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/votings/"+votingID.String()+"/votes", bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", votingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.CastVote)

	assert.Equal(t, http.StatusOK, rr.Code)

	select {
	case vote := <-voteChan:
		assert.Equal(t, votingID, vote.VotingID)
		assert.Equal(t, userID, vote.UserID)
		assert.Equal(t, domain.VoteYes, vote.Vote)
	case <-time.After(1 * time.Second):
		t.Error("expected vote to be sent to channel")
	}

	mockGroupRepo.AssertExpectations(t)
}

func TestVotingHandler_CastVote_VotingNotFound(t *testing.T) {
	_, _, mockGroupRepo, handler := setupVotingTest(t)

	userID := uuid.Must(uuid.NewV4())
	votingID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)

	reqBody := dto.CastVoteRequest{Vote: "yes"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/votings/"+votingID.String()+"/votes", bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", votingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.CastVote)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestVotingHandler_CastVote_VotingNotActive(t *testing.T) {
	votingService, _, mockGroupRepo, handler := setupVotingTest(t)

	userID := uuid.Must(uuid.NewV4())
	votingID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	voting := &domain.Voting{
		ID:      votingID,
		GroupID: groupID,
		Status:  domain.VotingStatusActive,
		EndsAt:  time.Now().Add(5 * time.Minute),
	}

	voteChan := make(chan *domain.Vote, 100)
	activeVoting := &service.ActiveVoting{
		Voting:   voting,
		VoteChan: voteChan,
		StopChan: make(chan struct{}),
		Votes:    make(map[uuid.UUID]domain.VoteType),
		Ctx:      ctx,
		Cancel:   cancel,
	}

	votingService.AddActiveVotingForTest(votingID.String(), activeVoting)

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)

	reqBody := dto.CastVoteRequest{Vote: "yes"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/votings/"+votingID.String()+"/votes", bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", votingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.CastVote)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestVotingHandler_CastVote_InvalidVoteType(t *testing.T) {
	_, _, _, handler := setupVotingTest(t)

	userID := uuid.Must(uuid.NewV4())
	votingID := uuid.Must(uuid.NewV4())

	reqBody := dto.CastVoteRequest{Vote: "invalid"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/votings/"+votingID.String()+"/votes", bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", votingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.CastVote)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

package service

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVotingService_CreateVoting_Success(t *testing.T) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	createdBy := uuid.Must(uuid.NewV4())
	duration := 5 * time.Minute

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, createdBy.String()).Return(domain.RoleAdmin, nil)
	mockVotingRepo.On("CreateVoting", mock.Anything, mock.AnythingOfType("*domain.Voting")).Return(nil)

	voting, err := service.CreateVoting(context.Background(), groupID, createdBy, "Feature", "Description", duration)

	assert.NoError(t, err)
	assert.Equal(t, "Feature", voting.FeatureName)
	assert.Equal(t, domain.VotingStatusActive, voting.Status)
	mockVotingRepo.AssertExpectations(t)
}

func TestVotingService_CreateVoting_NotAuthorized(t *testing.T) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	createdBy := uuid.Must(uuid.NewV4())
	duration := 5 * time.Minute

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, createdBy.String()).Return(domain.RoleMember, nil)

	voting, err := service.CreateVoting(context.Background(), groupID, createdBy, "Feature", "Description", duration)

	assert.Nil(t, voting)
	assert.ErrorIs(t, err, domain.ErrNotGroupOwner)
}

func TestVotingService_CastVote_Success(t *testing.T) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)

	votingID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	voting := &domain.Voting{
		ID:        uuid.Must(uuid.FromString(votingID)),
		GroupID:   groupID,
		CreatedBy: uuid.Must(uuid.NewV4()),
		Status:    domain.VotingStatusActive,
		EndsAt:    time.Now().Add(5 * time.Minute),
	}

	service.activeVotings.Store(votingID, &ActiveVoting{
		Voting:   voting,
		VoteChan: make(chan *domain.Vote, 100),
		StopChan: make(chan struct{}),
		Votes:    make(map[uuid.UUID]domain.VoteType),
		Ctx:      context.Background(),
		Cancel:   func() {},
	})

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)

	err := service.CastVote(context.Background(), votingID, userID, domain.VoteYes)

	assert.NoError(t, err)
}

func TestVotingService_GetVotingResult_FromActive(t *testing.T) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)

	votingID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	voting := &domain.Voting{
		ID:        uuid.Must(uuid.FromString(votingID)),
		GroupID:   groupID,
		CreatedBy: uuid.Must(uuid.NewV4()),
		Status:    domain.VotingStatusActive,
		EndsAt:    time.Now().Add(5 * time.Minute),
	}

	active := &ActiveVoting{
		Voting:   voting,
		VoteChan: make(chan *domain.Vote, 100),
		StopChan: make(chan struct{}),
		Votes: map[uuid.UUID]domain.VoteType{
			uuid.Must(uuid.NewV4()): domain.VoteYes,
			uuid.Must(uuid.NewV4()): domain.VoteNo,
		},
		Ctx:    context.Background(),
		Cancel: func() {},
	}
	service.activeVotings.Store(votingID, active)

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)
	mockGroupRepo.On("GetGroupMembersCount", mock.Anything, groupID.String()).Return(10, nil)

	result, err := service.GetVotingResult(context.Background(), votingID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.TotalVotes)
	assert.Equal(t, 1, result.YesVotes)
	assert.Equal(t, 1, result.NoVotes)
	assert.Equal(t, 10, result.TotalMembers)
}

func TestVotingService_GetVotingResult_FromDB(t *testing.T) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)

	votingID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	voting := &domain.Voting{
		ID:        uuid.Must(uuid.FromString(votingID)),
		GroupID:   groupID,
		CreatedBy: uuid.Must(uuid.NewV4()),
		Status:    domain.VotingStatusCompleted,
		EndsAt:    time.Now().Add(-5 * time.Minute),
	}

	mockVotingRepo.On("GetVotingByID", mock.Anything, votingID).Return(voting, nil)
	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)
	mockVotingRepo.On("CountVotesByType", mock.Anything, votingID).Return(5, 3, nil)
	mockVotingRepo.On("CountVotes", mock.Anything, votingID).Return(8, nil)
	mockGroupRepo.On("GetGroupMembersCount", mock.Anything, groupID.String()).Return(10, nil)

	result, err := service.GetVotingResult(context.Background(), votingID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 8, result.TotalVotes)
	assert.Equal(t, 5, result.YesVotes)
	assert.Equal(t, 3, result.NoVotes)
}

func TestVotingService_StopVoting_Success(t *testing.T) {
	mockVotingRepo := new(mocks.MockVotingRepository)
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewVotingService(mockVotingRepo, mockGroupRepo, mockUserRepo)

	votingID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())
	stopChan := make(chan struct{}, 1)
	active := &ActiveVoting{
		Voting: &domain.Voting{
			ID:        uuid.Must(uuid.FromString(votingID)),
			CreatedBy: userID,
		},
		StopChan: stopChan,
		Ctx:      context.Background(),
		Cancel:   func() {},
	}
	service.activeVotings.Store(votingID, active)

	err := service.StopVoting(context.Background(), votingID, userID)
	assert.NoError(t, err)

	select {
	case <-stopChan:
		// Signal recieved
	case <-time.After(100 * time.Millisecond):
		t.Error("expected stop signal to be sent to StopChan")
	}
}

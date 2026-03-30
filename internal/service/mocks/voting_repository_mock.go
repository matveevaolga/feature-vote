package mocks

import (
	"context"

	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockVotingRepository struct {
	mock.Mock
}

func (m *MockVotingRepository) CreateVoting(ctx context.Context, voting *domain.Voting) error {
	args := m.Called(ctx, voting)
	return args.Error(0)
}

func (m *MockVotingRepository) GetVotingByID(ctx context.Context, id string) (*domain.Voting, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Voting), args.Error(1)
}

func (m *MockVotingRepository) GetGroupVotings(ctx context.Context, groupID string, limit, offset int) ([]domain.Voting, error) {
	args := m.Called(ctx, groupID, limit, offset)
	return args.Get(0).([]domain.Voting), args.Error(1)
}

func (m *MockVotingRepository) UpdateVoting(ctx context.Context, voting *domain.Voting) error {
	args := m.Called(ctx, voting)
	return args.Error(0)
}

func (m *MockVotingRepository) CastVote(ctx context.Context, vote *domain.Vote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockVotingRepository) GetVote(ctx context.Context, votingID, userID string) (*domain.Vote, error) {
	args := m.Called(ctx, votingID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vote), args.Error(1)
}

func (m *MockVotingRepository) GetVotes(ctx context.Context, votingID string) ([]domain.Vote, error) {
	args := m.Called(ctx, votingID)
	return args.Get(0).([]domain.Vote), args.Error(1)
}

func (m *MockVotingRepository) CountVotes(ctx context.Context, votingID string) (int, error) {
	args := m.Called(ctx, votingID)
	return args.Int(0), args.Error(1)
}

func (m *MockVotingRepository) CountVotesByType(ctx context.Context, votingID string) (int, int, error) {
	args := m.Called(ctx, votingID)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockVotingRepository) GetGroupMembersCount(ctx context.Context, groupID string) (int, error) {
	args := m.Called(ctx, groupID)
	return args.Int(0), args.Error(1)
}

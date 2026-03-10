package repository

import (
	"context"

	"github.com/matveevaolga/feature-vote/internal/domain"
)

type VotingRepository interface {
	CreateVoting(ctx context.Context, voting *domain.Voting) error
	GetVotingByID(ctx context.Context, id string) (*domain.Voting, error)
	GetGroupVotings(ctx context.Context, groupID string, limit, offset int) ([]domain.Voting, error)
	UpdateVoting(ctx context.Context, voting *domain.Voting) error

	CastVote(ctx context.Context, vote *domain.Vote) error
	GetVote(ctx context.Context, votingID, userID string) (*domain.Vote, error)
	GetVotes(ctx context.Context, votingID string) ([]domain.Vote, error)
	CountVotes(ctx context.Context, votingID string) (int, error)
	CountVotesByType(ctx context.Context, votingID string) (yesCount, noCount int, err error)

	GetGroupMembersCount(ctx context.Context, groupID string) (int, error)
}

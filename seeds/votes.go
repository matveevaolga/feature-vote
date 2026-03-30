package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
)

func seedVotes(ctx context.Context, repo domainrepo.VotingRepository) error {
	votes, err := loadVotes()
	if err != nil {
		return err
	}

	for _, v := range votes {
		votingID, err := uuid.FromString(v.VotingID)
		if err != nil {
			return err
		}
		userID, err := uuid.FromString(v.UserID)
		if err != nil {
			return err
		}

		vote := &domain.Vote{
			VotingID:  votingID,
			UserID:    userID,
			Vote:      domain.VoteType(v.Vote),
			CreatedAt: v.CreatedAt,
		}

		if err := repo.CastVote(ctx, vote); err != nil {
			slog.Warn("failed to cast vote", "voting", v.VotingID, "user", v.UserID, "error", err)
			continue
		}
		slog.Info("vote cast", "voting", v.VotingID, "user", v.UserID)
	}
	return nil
}

func loadVotes() ([]VoteSeed, error) {
	data, err := seedFS.ReadFile("data/votes.json")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Votes []VoteSeed `json:"votes"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Votes, nil
}

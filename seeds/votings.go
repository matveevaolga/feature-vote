package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
)

func seedVotings(ctx context.Context, votingRepo domainrepo.VotingRepository, groupRepo domainrepo.GroupRepository) error {
	votings, err := loadVotings()
	if err != nil {
		return err
	}

	for _, v := range votings {
		id, err := uuid.FromString(v.ID)
		if err != nil {
			return err
		}
		groupID, err := uuid.FromString(v.GroupID)
		if err != nil {
			return err
		}
		createdBy, err := uuid.FromString(v.CreatedBy)
		if err != nil {
			return err
		}

		// Проверяем, существует ли группа
		if _, err := groupRepo.GetGroupByID(ctx, v.GroupID); err != nil {
			slog.Warn("skipping voting: group not found", "group", v.GroupID, "voting", v.ID)
			continue
		}

		voting := &domain.Voting{
			ID:          id,
			GroupID:     groupID,
			CreatedBy:   createdBy,
			FeatureName: v.FeatureName,
			Description: v.Description,
			Status:      domain.VotingStatus(v.Status),
			Duration:    time.Duration(v.Duration),
			StartedAt:   v.StartedAt,
			EndsAt:      v.EndsAt,
			CompletedAt: v.CompletedAt,
			Result:      v.Result,
		}

		if err := votingRepo.CreateVoting(ctx, voting); err != nil {
			slog.Warn("failed to create voting", "id", v.ID, "error", err)
			continue
		}
		slog.Info("voting created", "id", v.ID, "feature", v.FeatureName)
	}
	return nil
}

func loadVotings() ([]VotingSeed, error) {
	data, err := seedFS.ReadFile("data/votings.json")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Votings []VotingSeed `json:"votings"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Votings, nil
}

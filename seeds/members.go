package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
)

func seedMembers(ctx context.Context, repo domainrepo.GroupRepository) error {
	members, err := loadMembers()
	if err != nil {
		return err
	}

	for _, m := range members {
		groupID, err := uuid.FromString(m.GroupID)
		if err != nil {
			return err
		}
		userID, err := uuid.FromString(m.UserID)
		if err != nil {
			return err
		}

		member := &domain.GroupMember{
			GroupID:  groupID,
			UserID:   userID,
			Role:     domain.Role(m.Role),
			JoinedAt: m.JoinedAt,
		}

		if err := repo.AddMember(ctx, member); err != nil {
			slog.Warn("failed to add member", "group", m.GroupID, "user", m.UserID, "error", err)
			continue
		}
		slog.Info("member added", "group", m.GroupID, "user", m.UserID, "role", m.Role)
	}
	return nil
}

func loadMembers() ([]MemberSeed, error) {
	data, err := seedFS.ReadFile("data/members.json")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Members []MemberSeed `json:"members"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Members, nil
}

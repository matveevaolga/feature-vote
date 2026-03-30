package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
)

func seedGroups(ctx context.Context, repo domainrepo.GroupRepository) error {
	groups, err := loadGroups()
	if err != nil {
		return err
	}

	for _, g := range groups {
		groupID, err := uuid.FromString(g.ID)
		if err != nil {
			return err
		}

		ownerID, err := uuid.FromString(g.OwnerID)
		if err != nil {
			return err
		}

		group := &domain.Group{
			ID:        groupID,
			Name:      g.Name,
			OwnerID:   ownerID,
			CreatedAt: g.CreatedAt,
		}

		if err := repo.CreateGroup(ctx, group); err != nil {
			slog.Warn("failed to create group", "group", g.Name, "error", err)
			continue
		}
		slog.Info("group created", "name", group.Name, "id", group.ID)
	}
	return nil
}

func loadGroups() ([]GroupSeed, error) {
	data, err := seedFS.ReadFile("data/groups.json")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Groups []GroupSeed `json:"groups"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Groups, nil
}

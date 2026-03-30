package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
)

func seedInvitations(ctx context.Context, repo domainrepo.GroupRepository) error {
	invitations, err := loadInvitations()
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("invitations.json not found")
			return nil
		}
		return err
	}

	if len(invitations) == 0 {
		slog.Info("no invitations to seed")
		return nil
	}

	for _, inv := range invitations {
		id, err := uuid.FromString(inv.ID)
		if err != nil {
			slog.Warn("invalid invitation ID", "id", inv.ID, "error", err)
			continue
		}
		groupID, err := uuid.FromString(inv.GroupID)
		if err != nil {
			slog.Warn("invalid group ID in invitation", "group_id", inv.GroupID, "error", err)
			continue
		}
		userID, err := uuid.FromString(inv.UserID)
		if err != nil {
			slog.Warn("invalid user ID in invitation", "user_id", inv.UserID, "error", err)
			continue
		}

		invitation := &domain.Invitation{
			ID:        id,
			GroupID:   groupID,
			UserID:    userID,
			Status:    domain.Status(inv.Status),
			CreatedAt: inv.CreatedAt,
			UpdatedAt: inv.UpdatedAt,
		}

		if err := repo.CreateInvitation(ctx, invitation); err != nil {
			slog.Warn("failed to create invitation", "id", inv.ID, "error", err)
			continue
		}
		slog.Info("invitation created", "id", inv.ID, "group", inv.GroupID, "user", inv.UserID)
	}
	return nil
}

func loadInvitations() ([]InvitationSeed, error) {
	data, err := seedFS.ReadFile("data/invitations.json")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Invitations []InvitationSeed `json:"invitations"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Invitations, nil
}

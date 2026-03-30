package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
)

func seedUsers(ctx context.Context, repo domainrepo.UserRepository) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		exists, err := repo.GetUserByUsername(ctx, u.Username)
		if err != nil && err.Error() != "user not found" {
			return err
		}
		if exists != nil {
			slog.Info("user already exists", "username", u.Username)
			continue
		}

		userID, err := uuid.FromString(u.ID)
		if err != nil {
			return err
		}

		user := &domain.User{
			ID:        userID,
			Username:  u.Username,
			CreatedAt: u.CreatedAt,
		}

		if err := repo.CreateUser(ctx, user); err != nil {
			return err
		}
		slog.Info("user created", "username", user.Username, "id", user.ID)
	}
	return nil
}

func loadUsers() ([]UserSeed, error) {
	data, err := seedFS.ReadFile("data/users.json")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Users []UserSeed `json:"users"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Users, nil
}

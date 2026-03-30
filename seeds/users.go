package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	domainrepo "github.com/matveevaolga/feature-vote/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

func seedUsers(ctx context.Context, repo domainrepo.UserRepository) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		exists, err := repo.GetUserByEmail(ctx, u.Email)
		if err != nil && err != domain.ErrUserNotFound {
			return err
		}
		if exists != nil {
			slog.Info("user already exists", "email", u.Email)
			continue
		}

		userID, err := uuid.FromString(u.ID)
		if err != nil {
			return err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user := &domain.User{
			ID:           userID,
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: string(hashedPassword),
			CreatedAt:    u.CreatedAt,
			UpdatedAt:    u.UpdatedAt,
		}

		if err := repo.CreateUser(ctx, user); err != nil {
			return err
		}
		slog.Info("user created", "username", user.Username, "email", user.Email)
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

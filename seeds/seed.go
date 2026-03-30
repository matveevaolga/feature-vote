package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matveevaolga/feature-vote/internal/config"
	"github.com/matveevaolga/feature-vote/internal/logger"
	"github.com/matveevaolga/feature-vote/internal/repository"
)

//go:embed data/*
var seedFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger.Init(cfg.LogLevel)
	slog.Info("Starting seed data loader")

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	votingRepo := repository.NewVotingRepository(db)

	if err := seedUsers(ctx, userRepo); err != nil {
		slog.Error("Failed to seed users", "error", err)
		os.Exit(1)
	}

	if err := seedGroups(ctx, groupRepo); err != nil {
		slog.Error("Failed to seed groups", "error", err)
		os.Exit(1)
	}

	if err := seedMembers(ctx, groupRepo); err != nil {
		slog.Error("Failed to seed members", "error", err)
		os.Exit(1)
	}

	if err := seedInvitations(ctx, groupRepo); err != nil {
		slog.Error("Failed to seed invitations", "error", err)
		os.Exit(1)
	}

	if err := seedVotings(ctx, votingRepo, groupRepo); err != nil {
		slog.Error("Failed to seed votings", "error", err)
		os.Exit(1)
	}

	if err := seedVotes(ctx, votingRepo); err != nil {
		slog.Error("Failed to seed votes", "error", err)
		os.Exit(1)
	}

	slog.Info("Seed completed successfully")
}

func checkTableEmpty(ctx context.Context, db *pgxpool.Pool, tableName string) (bool, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

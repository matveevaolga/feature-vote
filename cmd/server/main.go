package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/matveevaolga/feature-vote/internal/config"
	"github.com/matveevaolga/feature-vote/internal/logger"
	"github.com/matveevaolga/feature-vote/internal/repository"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

func registerHttpRoutes(
	userHandler *handler.UserHandler,
	groupHandler *handler.GroupHandler,
	votingHandler *handler.VotingHandler,
) {
	// User endpoints
	http.HandleFunc("POST /users", middleware.Logging(userHandler.CreateUser))

	// Group endpoints
	http.HandleFunc("POST /groups", middleware.Logging(middleware.Auth(groupHandler.CreateGroup)))
	http.HandleFunc("GET /groups/{id}", middleware.Logging(middleware.Auth(groupHandler.GetGroup)))
	http.HandleFunc("PUT /groups/{id}", middleware.Logging(middleware.Auth(groupHandler.UpdateGroup)))
	http.HandleFunc("DELETE /groups/{id}", middleware.Logging(middleware.Auth(groupHandler.DeleteGroup)))
	http.HandleFunc("GET /users/groups", middleware.Logging(middleware.Auth(groupHandler.GetUserGroups)))

	// Member endpoints
	http.HandleFunc("POST /groups/{id}/invite", middleware.Logging(middleware.Auth(groupHandler.InviteMember)))
	http.HandleFunc("GET /groups/{id}/members", middleware.Logging(middleware.Auth(groupHandler.GetGroupMembers)))
	http.HandleFunc("DELETE /groups/{id}/members/{userID}", middleware.Logging(middleware.Auth(groupHandler.RemoveMember)))
	http.HandleFunc("POST /groups/{id}/leave", middleware.Logging(middleware.Auth(groupHandler.LeaveGroup)))
	http.HandleFunc("PUT /groups/{id}/members/{userID}/role", middleware.Logging(middleware.Auth(groupHandler.UpdateMemberRole)))
	http.HandleFunc("POST /groups/{id}/transfer", middleware.Logging(middleware.Auth(groupHandler.TransferOwnership)))

	// Invitation endpoints
	http.HandleFunc("GET /users/invitations", middleware.Logging(middleware.Auth(groupHandler.GetInvitations)))
	http.HandleFunc("POST /invitations/{id}/accept", middleware.Logging(middleware.Auth(groupHandler.AcceptInvitation)))
	http.HandleFunc("POST /invitations/{id}/decline", middleware.Logging(middleware.Auth(groupHandler.DeclineInvitation)))

	// Voting endpoints
	http.HandleFunc("POST /votings", middleware.Logging(middleware.Auth(votingHandler.CreateVoting)))
	http.HandleFunc("GET /votings/{id}", middleware.Logging(middleware.Auth(votingHandler.GetVotingStatus)))
	http.HandleFunc("GET /votings/{id}/results", middleware.Logging(middleware.Auth(votingHandler.GetVotingResults)))
	http.HandleFunc("POST /votings/{id}/votes", middleware.Logging(middleware.Auth(votingHandler.CastVote)))
	http.HandleFunc("POST /votings/{id}/stop", middleware.Logging(middleware.Auth(votingHandler.StopVoting)))
}

func main() {
	logger.Init("info")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	votingRepo := repository.NewVotingRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)
	votingService := service.NewVotingService(votingRepo, groupRepo, userRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	groupHandler := handler.NewGroupHandler(groupService, userService)
	votingHandler := handler.NewVotingHandler(votingService)

	// Register all routes
	registerHttpRoutes(userHandler, groupHandler, votingHandler)

	slog.Info("Server starting", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/matveevaolga/feature-vote/internal/config"
	"github.com/matveevaolga/feature-vote/internal/logger"
	"github.com/matveevaolga/feature-vote/internal/repository"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

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

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	votingRepo := repository.NewVotingRepository(db)

	userService := service.NewUserService(userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)
	votingService := service.NewVotingService(votingRepo, groupRepo, userRepo)

	userHandler := handler.NewUserHandler(userService)
	groupHandler := handler.NewGroupHandler(groupService, userService)
	votingHandler := handler.NewVotingHandler(votingService)

	r := chi.NewRouter()
	r.Use(chi_middleware.RequestID)
	r.Use(chi_middleware.RealIP)
	r.Use(chi_middleware.Recoverer)
	r.Use(chi_middleware.Timeout(60 * time.Second))
	r.Post("/users", middleware.Logging(http.HandlerFunc(userHandler.CreateUser)).ServeHTTP)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(middleware.Logging)
		r.Route("/groups", func(r chi.Router) {
			r.Post("/", groupHandler.CreateGroup)
			r.Get("/{id}", groupHandler.GetGroup)
			r.Put("/{id}", groupHandler.UpdateGroup)
			r.Delete("/{id}", groupHandler.DeleteGroup)
			r.Route("/{id}/members", func(r chi.Router) {
				r.Get("/", groupHandler.GetGroupMembers)
				r.Post("/invite", groupHandler.InviteMember)
				r.Delete("/{userID}", groupHandler.RemoveMember)
				r.Put("/{userID}/role", groupHandler.UpdateMemberRole)
			})
			r.Post("/{id}/leave", groupHandler.LeaveGroup)
			r.Post("/{id}/transfer", groupHandler.TransferOwnership)
		})
		r.Route("/votings", func(r chi.Router) {
			r.Post("/", votingHandler.CreateVoting)
			r.Get("/{id}", votingHandler.GetVotingStatus)
			r.Get("/{id}/results", votingHandler.GetVotingResults)
			r.Post("/{id}/votes", votingHandler.CastVote)
			r.Post("/{id}/stop", votingHandler.StopVoting)
		})
		r.Get("/users/groups", groupHandler.GetUserGroups)
		r.Get("/users/invitations", groupHandler.GetInvitations)
		r.Post("/invitations/{id}/accept", groupHandler.AcceptInvitation)
		r.Post("/invitations/{id}/decline", groupHandler.DeclineInvitation)
	})

	slog.Info("Server starting", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

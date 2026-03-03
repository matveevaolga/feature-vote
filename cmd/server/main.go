package main

import (
	"log"
	"net/http"

	"github.com/matveevaolga/feature-vote/internal/config"
	"github.com/matveevaolga/feature-vote/internal/repository"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	http.HandleFunc("/users", userHandler.CreateUser)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatal("Server failed:", err)
	}
}

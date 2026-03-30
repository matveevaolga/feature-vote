package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
)

type HealthHandler struct {
	healthService *service.HealthService
	startTime     time.Time
}

func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
		startTime:     time.Now(),
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "up"
	if err := h.healthService.CheckDatabase(ctx); err != nil {
		dbStatus = "down"
	}

	votingsCount := h.healthService.GetActiveVotingsCount()
	votingsStatus := "ok"
	if votingsCount > 0 {
		votingsStatus = "active"
	}

	response := dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks: dto.HealthChecks{
			Database: dbStatus,
			Votings:  votingsStatus,
			Uptime:   time.Since(h.startTime).String(),
		},
	}

	if dbStatus == "down" {
		response.Status = "unhealthy"
		RespondWithJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	RespondWithJSON(w, http.StatusOK, response)
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.healthService.CheckDatabase(ctx); err != nil {
		RespondWithJSON(w, http.StatusServiceUnavailable, dto.HealthErrorResponse{
			Status:    "not ready",
			Timestamp: time.Now(),
			Error:     "database connection failed",
		})
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
	})
}

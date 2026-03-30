package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthService struct {
	db            *pgxpool.Pool
	votingService *VotingService
}

func NewHealthService(db *pgxpool.Pool, votingService *VotingService) *HealthService {
	return &HealthService{
		db:            db,
		votingService: votingService,
	}
}

func (s *HealthService) CheckDatabase(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *HealthService) GetActiveVotingsCount() int {
	count := 0
	s.votingService.activeVotings.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

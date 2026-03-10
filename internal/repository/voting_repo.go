package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matveevaolga/feature-vote/internal/domain"
)

type votingRepository struct {
	db *pgxpool.Pool
}

func NewVotingRepository(db *pgxpool.Pool) *votingRepository {
	return &votingRepository{db: db}
}

func (r *votingRepository) CreateVoting(ctx context.Context, voting *domain.Voting) error {
	query := `
		INSERT INTO votings (id, group_id, created_by, feature_name, description, 
			status, duration, started_at, ends_at, completed_at, result) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		voting.ID, voting.GroupID, voting.CreatedBy, voting.FeatureName, voting.Description,
		voting.Status, voting.Duration, voting.StartedAt, voting.EndsAt, voting.CompletedAt, voting.Result,
	)

	if err != nil {
		return fmt.Errorf("failed to create voting: %w", err)
	}
	return nil
}

func (r *votingRepository) GetVotingByID(ctx context.Context, id string) (*domain.Voting, error) {
	query := `SELECT id, group_id, created_by, feature_name, description, 
	status, duration, started_at, ends_at, completed_at, result FROM votings WHERE id = $1`
	var voting domain.Voting
	err := r.db.QueryRow(ctx, query, id).Scan(
		&voting.ID, &voting.GroupID, &voting.CreatedBy, &voting.FeatureName, &voting.Description,
		&voting.Status, &voting.Duration, &voting.StartedAt, &voting.EndsAt, &voting.CompletedAt, &voting.Result,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVotingNotFound
		}
		return nil, fmt.Errorf("failed to get voting by id: %w", err)
	}
	return &voting, nil
}

func (r *votingRepository) GetGroupVotings(ctx context.Context, groupID string, limit, offset int) ([]domain.Voting, error) {
	query := `
        SELECT id, group_id, created_by, feature_name, description, status, duration, started_at, ends_at, completed_at, result 
        FROM votings WHERE group_id = $1 ORDER BY started_at DESC LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get group votings: %w", err)
	}
	defer rows.Close()
	var votings []domain.Voting
	for rows.Next() {
		var voting domain.Voting
		err := rows.Scan(
			&voting.ID, &voting.GroupID, &voting.CreatedBy, &voting.FeatureName, &voting.Description,
			&voting.Status, &voting.Duration, &voting.StartedAt, &voting.EndsAt, &voting.CompletedAt, &voting.Result,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan voting: %w", err)
		}
		votings = append(votings, voting)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}
	return votings, nil
}

func (r *votingRepository) UpdateVoting(ctx context.Context, voting *domain.Voting) error {
	query := `
		UPDATE votings 
		SET status = $1, completed_at = $2, result = $3
		WHERE id = $4
	`
	commandTag, err := r.db.Exec(ctx, query,
		voting.Status, voting.CompletedAt, voting.Result, voting.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update voting: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrVotingNotFound
	}
	return nil
}

func (r *votingRepository) CastVote(ctx context.Context, vote *domain.Vote) error {
	var status string
	checkQuery := `SELECT status FROM votings WHERE id = $1`
	err := r.db.QueryRow(ctx, checkQuery, vote.VotingID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVotingNotFound
		}
		return fmt.Errorf("failed to check voting status: %w", err)
	}
	if domain.VotingStatus(status) != domain.VotingStatusActive {
		return domain.ErrVotingNotActive
	}

	var exists bool
	duplicateQuery := `SELECT EXISTS(SELECT 1 FROM votes WHERE voting_id = $1 AND user_id = $2)`
	err = r.db.QueryRow(ctx, duplicateQuery, vote.VotingID, vote.UserID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check duplicate vote: %w", err)
	}
	if exists {
		return domain.ErrAlreadyVoted
	}

	query := `
		INSERT INTO votes (voting_id, user_id, vote, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = r.db.Exec(ctx, query, vote.VotingID, vote.UserID, vote.Vote, vote.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to cast vote: %w", err)
	}
	return nil
}

func (r *votingRepository) GetVote(ctx context.Context, votingID, userID string) (*domain.Vote, error) {
	query := `
		SELECT voting_id, user_id, vote, created_at
		FROM votes
		WHERE voting_id = $1 AND user_id = $2
	`
	var vote domain.Vote
	err := r.db.QueryRow(ctx, query, votingID, userID).Scan(
		&vote.VotingID, &vote.UserID, &vote.Vote, &vote.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVoteNotFound
		}
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	return &vote, nil
}

func (r *votingRepository) GetVotes(ctx context.Context, votingID string) ([]domain.Vote, error) {
	query := `
        SELECT voting_id, user_id, vote, created_at FROM votes WHERE voting_id = $1 ORDER BY created_at DESC
    `
	rows, err := r.db.Query(ctx, query, votingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes: %w", err)
	}
	defer rows.Close()
	var votes []domain.Vote
	for rows.Next() {
		var vote domain.Vote
		err := rows.Scan(&vote.VotingID, &vote.UserID, &vote.Vote, &vote.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votes = append(votes, vote)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}
	return votes, nil
}

func (r *votingRepository) CountVotes(ctx context.Context, votingID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM votes WHERE voting_id = $1`
	err := r.db.QueryRow(ctx, query, votingID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count votes: %w", err)
	}
	return count, nil
}

func (r *votingRepository) CountVotesByType(ctx context.Context, votingID string) (int, int, error) {
	query := `SELECT COUNT(CASE WHEN vote = 'yes' THEN 1 END) as yes_count,
			COUNT(CASE WHEN vote = 'no' THEN 1 END) as no_count
			FROM votes WHERE voting_id = $1
	`
	var yesCount, noCount int
	err := r.db.QueryRow(ctx, query, votingID).Scan(&yesCount, &noCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count votes by type: %w", err)
	}

	return yesCount, noCount, nil
}

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
)

func (r *groupRepository) CreateInvitation(ctx context.Context, inv *domain.Invitation) error {
	checkQuery := `SELECT id FROM invitations WHERE group_id = $1 AND user_id = $2 AND status = $3`
	var existID string
	err := r.db.QueryRow(ctx, checkQuery, inv.GroupID, inv.UserID, domain.StatusPending).Scan(&existID)
	if err == nil {
		return domain.ErrInvitationAlreadySent
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check existing invitation: %w", err)
	}

	query := `INSERT INTO invitations (id, group_id, user_id, status, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = r.db.Exec(ctx, query, inv.ID, inv.GroupID, inv.UserID, inv.Status, inv.CreatedAt, inv.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}
	return nil
}

func (r *groupRepository) GetInvitation(ctx context.Context, id string) (*domain.Invitation, error) {
	var inv domain.Invitation
	var statusString string
	query := `SELECT id, group_id, user_id, status, created_at, updated_at FROM invitations WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&inv.ID, &inv.GroupID, &inv.UserID, &statusString, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	inv.Status = domain.Status(statusString)
	return &inv, nil
}

func (r *groupRepository) GetInvitationWithGroup(ctx context.Context, id string) (*domain.InvitationWithGroup, error) {
	var inv domain.InvitationWithGroup
	var statusString string
	query := `
		SELECT i.id, i.group_id, i.user_id, i.status, i.created_at, i.updated_at, g.name as group_name
		FROM invitations i
		JOIN groups g ON i.group_id = g.id
		WHERE i.id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.GroupID, &inv.UserID, &statusString, &inv.CreatedAt, &inv.UpdatedAt, &inv.GroupName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get invitation with group: %w", err)
	}
	inv.Status = domain.Status(statusString)
	return &inv, nil
}

func (r *groupRepository) GetUserInvitationsWithGroup(ctx context.Context, userID string) ([]domain.InvitationWithGroup, error) {
	query := `
		SELECT i.id, i.group_id, i.user_id, i.status, i.created_at, i.updated_at, g.name as group_name
		FROM invitations i
		JOIN groups g ON i.group_id = g.id
		WHERE i.user_id = $1 AND i.status = $2
		ORDER BY i.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID, domain.StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get user invitations: %w", err)
	}
	defer rows.Close()

	var invitations []domain.InvitationWithGroup
	for rows.Next() {
		var inv domain.InvitationWithGroup
		var statusString string
		err := rows.Scan(&inv.ID, &inv.GroupID, &inv.UserID, &statusString, &inv.CreatedAt, &inv.UpdatedAt, &inv.GroupName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		inv.Status = domain.Status(statusString)
		invitations = append(invitations, inv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}

	return invitations, nil
}

func (r *groupRepository) UpdateInvitation(ctx context.Context, inv *domain.Invitation) error {
	query := `UPDATE invitations SET status = $1, updated_at = $2 WHERE id = $3 AND status = $4`
	commandTag, err := r.db.Exec(ctx, query, inv.Status, inv.UpdatedAt, inv.ID, domain.StatusPending)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		var status string
		checkQuery := `SELECT status FROM invitations WHERE id = $1`
		err := r.db.QueryRow(ctx, checkQuery, inv.ID).Scan(&status)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrInvitationNotFound
			}
			return fmt.Errorf("failed to check invitation: %w", err)
		}
		return domain.ErrInvitationNotPending
	}
	return nil
}

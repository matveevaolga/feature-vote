package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matveevaolga/feature-vote/internal/domain"
)

type groupRepository struct {
	db *pgxpool.Pool
}

func NewGroupRepository(db *pgxpool.Pool) *groupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) CreateGroup(ctx context.Context, group *domain.Group) error {
	query := `INSERT INTO groups (id, name, owner_id, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, group.ID, group.Name, group.OwnerID, group.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (r *groupRepository) GetGroupByID(ctx context.Context, id string) (*domain.Group, error) {
	query := `SELECT id, name, owner_id, created_at FROM groups WHERE id = $1`
	var group domain.Group
	err := r.db.QueryRow(ctx, query, id).Scan(&group.ID, &group.Name, &group.OwnerID, &group.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group by id: %w", err)
	}
	return &group, nil
}

func (r *groupRepository) UpdateGroup(ctx context.Context, group *domain.Group) error {
	query := `UPDATE groups SET name = $1 WHERE id = $2`
	commandTag, err := r.db.Exec(ctx, query, group.Name, group.ID)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

func (r *groupRepository) DeleteGroup(ctx context.Context, id string) error {
	query := `DELETE FROM groups WHERE id = $1`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

func (r *groupRepository) GetUserGroups(ctx context.Context, userID string) ([]domain.Group, error) {
	query := `
		SELECT g.id, g.name, g.owner_id, g.created_at
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		var group domain.Group
		err := rows.Scan(&group.ID, &group.Name, &group.OwnerID, &group.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}
	return groups, nil
}

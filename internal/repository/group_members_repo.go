package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
)

func (r *groupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	query := `INSERT INTO group_members (group_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, member.GroupID, member.UserID, member.Role, member.JoinedAt)
	if err != nil {
		return fmt.Errorf("failed to add member to group: %w", err)
	}
	return nil
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	var role string
	roleCheckQuery := `SELECT role FROM group_members WHERE group_id = $1 AND user_id = $2`
	err := r.db.QueryRow(ctx, roleCheckQuery, groupID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotGroupMember
		}
		return fmt.Errorf("failed to check member role: %w", err)
	}

	if role == string(domain.RoleOwner) {
		return fmt.Errorf("cannot remove group owner")
	}

	query := `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`
	commandTag, err := r.db.Exec(ctx, query, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member from group: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrNotGroupMember
	}
	return nil
}

func (r *groupRepository) GetMembers(ctx context.Context, groupID string) ([]domain.GroupMember, error) {
	query := `SELECT group_id, user_id, role, joined_at FROM group_members WHERE group_id = $1 ORDER BY joined_at DESC`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []domain.GroupMember
	for rows.Next() {
		var member domain.GroupMember
		var roleString string
		err := rows.Scan(&member.GroupID, &member.UserID, &roleString, &member.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		member.Role = domain.Role(roleString)
		members = append(members, member)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}
	return members, nil
}

func (r *groupRepository) GetMemberRole(ctx context.Context, groupID, userID string) (domain.Role, error) {
	query := `SELECT role FROM group_members WHERE group_id = $1 AND user_id = $2`
	var roleString string
	err := r.db.QueryRow(ctx, query, groupID, userID).Scan(&roleString)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotGroupMember
		}
		return "", fmt.Errorf("failed to get member role: %w", err)
	}
	return domain.Role(roleString), nil
}

func (r *groupRepository) UpdateMemberRole(ctx context.Context, groupID, userID string, newRole domain.Role) error {
	currentRole, err := r.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if currentRole == domain.RoleOwner {
		return fmt.Errorf("cannot change owner's role")
	}
	if !newRole.Valid() {
		return fmt.Errorf("invalid role: %s", newRole)
	}
	query := `UPDATE group_members SET role = $1 WHERE group_id = $2 AND user_id = $3`
	commandTag, err := r.db.Exec(ctx, query, newRole, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrNotGroupMember
	}
	return nil
}

func (r *groupRepository) GetGroupMembersCount(ctx context.Context, groupID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_members WHERE group_id = $1`
	err := r.db.QueryRow(ctx, query, groupID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count group members: %w", err)
	}
	return count, nil
}

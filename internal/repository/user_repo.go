package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matveevaolga/feature-vote/internal/domain"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.CreatedAt)
	return err
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, created_at FROM users WHERE username = $1`
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	query := `SELECT id, username, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user domain.User
		err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET username = $1 WHERE id = $2`
	commandTag, err := r.db.Exec(ctx, query, user.Username, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

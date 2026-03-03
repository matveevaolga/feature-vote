package service

import (
	"context"
	"errors"

	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/domain/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	existingUser, err := s.userRepo.GetUserByUsername(ctx, user.Username)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	if existingUser != nil {
		return domain.ErrUserAlreadyExists
	}

	return s.userRepo.CreateUser(ctx, user)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetUserByUsername(ctx, username)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	return s.userRepo.ListUsers(ctx, limit, offset)
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	existingUser, err := s.userRepo.GetUserByID(ctx, user.ID.String())
	if err != nil {
		return err
	}

	if existingUser.Username != user.Username {
		userWithSameName, err := s.userRepo.GetUserByUsername(ctx, user.Username)
		if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
			return err
		}
		if userWithSameName != nil {
			return domain.ErrUserAlreadyExists
		}
	}

	return s.userRepo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.DeleteUser(ctx, id)
}

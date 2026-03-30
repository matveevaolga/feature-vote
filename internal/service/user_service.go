package service

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo   repository.UserRepository
	jwtService JWTServiceInterface
}

func NewUserService(userRepo repository.UserRepository, jwtService JWTServiceInterface) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type RegisterParams struct {
	Username string
	Email    string
	Password string
}

type LoginParams struct {
	Email    string
	Password string
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	if user.Email == "" {
		user.Email = user.Username + "@temp.local"
	}
	if user.PasswordHash == "" {
		user.PasswordHash = "temporary_hash_for_imported_users"
	}
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
	existingUser, err = s.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	if existingUser != nil {
		return domain.ErrUserAlreadyExists
	}
	user.UpdatedAt = time.Now()
	return s.userRepo.CreateUser(ctx, user)
}

func (s *UserService) Register(ctx context.Context, params RegisterParams) (*domain.User, error) {
	if len(params.Username) < 5 || len(params.Username) > 50 {
		return nil, domain.ErrInvalidUsername
	}
	if len(params.Email) < 5 || len(params.Email) > 100 {
		return nil, domain.ErrInvalidEmail
	}
	if len(params.Password) < 6 {
		return nil, domain.ErrInvalidPassword
	}

	existing, _ := s.userRepo.GetUserByUsername(ctx, params.Username)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	existing, _ = s.userRepo.GetUserByEmail(ctx, params.Email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.Must(uuid.NewV4()),
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, params LoginParams) (*domain.User, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetUserByUsername(ctx, username)
}

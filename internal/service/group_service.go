package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/domain/repository"
)

type GroupService struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
}

func NewGroupService(groupRepo repository.GroupRepository, userRepo repository.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, name string, ownerID uuid.UUID) (*domain.Group, error) {
	if !domain.ValidateGroupName(name) {
		return nil, domain.ErrInvalidGroupName
	}
	_, err := s.userRepo.GetUserByID(ctx, ownerID.String())
	if err != nil {
		return nil, err
	}
	group, owner := domain.NewGroup(name, ownerID)
	if err := s.groupRepo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}
	if err := s.groupRepo.AddMember(ctx, owner); err != nil {
		s.groupRepo.DeleteGroup(ctx, group.ID.String())
		return nil, err
	}
	return group, nil
}

func (s *GroupService) GetGroupByID(ctx context.Context, groupID string) (*domain.Group, error) {
	return s.groupRepo.GetGroupByID(ctx, groupID)
}

func (s *GroupService) UpdateGroup(ctx context.Context, groupID, name string, ownerID uuid.UUID) error {
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if !group.IsOwner(ownerID) {
		return domain.ErrNotGroupOwner
	}
	if !domain.ValidateGroupName(name) {
		return domain.ErrInvalidGroupName
	}
	group.Name = name
	return s.groupRepo.UpdateGroup(ctx, group)
}

func (s *GroupService) DeleteGroup(ctx context.Context, groupID string, ownerID uuid.UUID) error {
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if !group.IsOwner(ownerID) {
		return domain.ErrNotGroupOwner
	}
	return s.groupRepo.DeleteGroup(ctx, groupID)
}

func (s *GroupService) GetUserGroups(ctx context.Context, userID string) ([]domain.Group, error) {
	return s.groupRepo.GetUserGroups(ctx, userID)
}

func (s *GroupService) InviteMember(ctx context.Context, groupID string, userID, ownerID uuid.UUID) error {
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, ownerID.String())
	if err != nil {
		return err
	}
	if !role.CanManageMembers() {
		return domain.ErrNotGroupOwner
	}
	_, err = s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return err
	}
	_, err = s.groupRepo.GetMemberRole(ctx, groupID, userID.String())
	if err == nil {
		return domain.ErrAlreadyGroupMember
	}
	if !errors.Is(err, domain.ErrNotGroupMember) {
		return err
	}
	inv := &domain.Invitation{
		ID:        uuid.Must(uuid.NewV4()),
		GroupID:   uuid.Must(uuid.FromString(groupID)),
		UserID:    userID,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.groupRepo.CreateInvitation(ctx, inv)
}

func (s *GroupService) GetInvitation(ctx context.Context, invitationID string, userID uuid.UUID) (*domain.InvitationWithGroup, error) {
	inv, err := s.groupRepo.GetInvitationWithGroup(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if inv.UserID != userID {
		return nil, domain.ErrInvitationNotFound
	}
	return inv, nil
}

func (s *GroupService) AcceptInvitation(ctx context.Context, invitationID string, userID uuid.UUID) error {
	inv, err := s.groupRepo.GetInvitation(ctx, invitationID)
	if err != nil {
		return err
	}
	if inv.UserID != userID {
		return domain.ErrInvitationNotFound
	}
	if !inv.IsPending() {
		return domain.ErrInvitationNotPending
	}
	if err := inv.Accept(); err != nil {
		return err
	}
	member := &domain.GroupMember{
		GroupID:  inv.GroupID,
		UserID:   inv.UserID,
		Role:     domain.RoleMember,
		JoinedAt: time.Now(),
	}
	if err := s.groupRepo.AddMember(ctx, member); err != nil {
		return err
	}
	return s.groupRepo.UpdateInvitation(ctx, inv)
}

func (s *GroupService) DeclineInvitation(ctx context.Context, invitationID string, userID uuid.UUID) error {
	inv, err := s.groupRepo.GetInvitation(ctx, invitationID)
	if err != nil {
		return err
	}
	if inv.UserID != userID {
		return domain.ErrInvitationNotFound
	}
	if !inv.IsPending() {
		return domain.ErrInvitationNotPending
	}
	if err := inv.Decline(); err != nil {
		return err
	}
	return s.groupRepo.UpdateInvitation(ctx, inv)
}

func (s *GroupService) GetUserInvitations(ctx context.Context, userID string) ([]domain.InvitationWithGroup, error) {
	return s.groupRepo.GetUserInvitationsWithGroup(ctx, userID)
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID string, ownerID, userID uuid.UUID) error {
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, ownerID.String())
	if err != nil {
		return err
	}
	if role != domain.RoleOwner {
		return domain.ErrNotGroupOwner
	}
	if ownerID == userID {
		return domain.ErrCannotRemoveGroupOwner
	}
	return s.groupRepo.RemoveMember(ctx, groupID, userID.String())
}

func (s *GroupService) LeaveGroup(ctx context.Context, groupID string, userID uuid.UUID) error {
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, userID.String())
	if err != nil {
		return err
	}
	if role == domain.RoleOwner {
		return domain.ErrCannotRemoveGroupOwner
	}
	return s.groupRepo.RemoveMember(ctx, groupID, userID.String())
}

func (s *GroupService) GetGroupMembers(ctx context.Context, groupID string, userID uuid.UUID) ([]domain.GroupMember, error) {
	_, err := s.groupRepo.GetMemberRole(ctx, groupID, userID.String())
	if err != nil {
		return nil, err
	}
	return s.groupRepo.GetMembers(ctx, groupID)
}

func (s *GroupService) UpdateMemberRole(ctx context.Context, groupID string, ownerID, userID uuid.UUID, newRole domain.Role) error {
	if !newRole.Valid() {
		return fmt.Errorf("invalid role: %s", newRole)
	}
	ownerRole, err := s.groupRepo.GetMemberRole(ctx, groupID, ownerID.String())
	if err != nil {
		return err
	}
	if ownerRole != domain.RoleOwner {
		return domain.ErrNotGroupOwner
	}
	if ownerID == userID {
		return fmt.Errorf("cannot change your own role as owner")
	}
	return s.groupRepo.UpdateMemberRole(ctx, groupID, userID.String(), newRole)
}

func (s *GroupService) TransferOwnership(ctx context.Context, groupID string, currentOwnerID, newOwnerID uuid.UUID) error {
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if !group.IsOwner(currentOwnerID) {
		return domain.ErrNotGroupOwner
	}
	_, err = s.groupRepo.GetMemberRole(ctx, groupID, newOwnerID.String())
	if err != nil {
		return err
	}
	if currentOwnerID == newOwnerID {
		return fmt.Errorf("new owner must be different from current owner")
	}
	group.OwnerID = newOwnerID
	if err := s.groupRepo.UpdateGroup(ctx, group); err != nil {
		return err
	}
	if err := s.groupRepo.UpdateMemberRole(ctx, groupID, currentOwnerID.String(), domain.RoleAdmin); err != nil {
		return err
	}
	if err := s.groupRepo.UpdateMemberRole(ctx, groupID, newOwnerID.String(), domain.RoleOwner); err != nil {
		return err
	}
	return nil
}

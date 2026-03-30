package service

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGroupService_CreateGroup_Success(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	ownerID := uuid.Must(uuid.NewV4())
	groupName := "Test Group"

	mockUserRepo.On("GetUserByID", mock.Anything, ownerID.String()).Return(&domain.User{ID: ownerID}, nil)
	mockGroupRepo.On("CreateGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
	mockGroupRepo.On("AddMember", mock.Anything, mock.AnythingOfType("*domain.GroupMember")).Return(nil)

	group, err := service.CreateGroup(context.Background(), groupName, ownerID)

	assert.NoError(t, err)
	assert.Equal(t, groupName, group.Name)
	assert.Equal(t, ownerID, group.OwnerID)
	mockGroupRepo.AssertExpectations(t)
}

func TestGroupService_CreateGroup_InvalidName(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	ownerID := uuid.Must(uuid.NewV4())

	group, err := service.CreateGroup(context.Background(), "ab", ownerID)

	assert.Nil(t, group)
	assert.ErrorIs(t, err, domain.ErrInvalidGroupName)
}

func TestGroupService_CreateGroup_UserNotFound(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	ownerID := uuid.Must(uuid.NewV4())

	mockUserRepo.On("GetUserByID", mock.Anything, ownerID.String()).Return(nil, domain.ErrUserNotFound)

	group, err := service.CreateGroup(context.Background(), "Test Group", ownerID)

	assert.Nil(t, group)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestGroupService_InviteMember_Success(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	inviterID := uuid.Must(uuid.NewV4())
	inviteeID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, inviterID.String()).Return(domain.RoleOwner, nil)
	mockUserRepo.On("GetUserByID", mock.Anything, inviteeID.String()).Return(&domain.User{ID: inviteeID}, nil)
	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, inviteeID.String()).Return(domain.Role(""), domain.ErrNotGroupMember)
	mockGroupRepo.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.Invitation")).Return(nil)

	err := service.InviteMember(context.Background(), groupID, inviteeID, inviterID)

	assert.NoError(t, err)
	mockGroupRepo.AssertExpectations(t)
}

func TestGroupService_InviteMember_NotOwner(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	inviterID := uuid.Must(uuid.NewV4())
	inviteeID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, inviterID.String()).Return(domain.RoleMember, nil)

	err := service.InviteMember(context.Background(), groupID, inviteeID, inviterID)

	assert.ErrorIs(t, err, domain.ErrNotGroupOwner)
}

func TestGroupService_AcceptInvitation_Success(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	invitationID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	invitation := &domain.Invitation{
		ID:      uuid.Must(uuid.FromString(invitationID)),
		GroupID: groupID,
		UserID:  userID,
		Status:  domain.StatusPending,
	}

	mockGroupRepo.On("GetInvitation", mock.Anything, invitationID).Return(invitation, nil)
	mockGroupRepo.On("AddMember", mock.Anything, mock.AnythingOfType("*domain.GroupMember")).Return(nil)
	mockGroupRepo.On("UpdateInvitation", mock.Anything, invitation).Return(nil)

	err := service.AcceptInvitation(context.Background(), invitationID, userID)

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusAccepted, invitation.Status)
}

func TestGroupService_AcceptInvitation_NotFound(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	invitationID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetInvitation", mock.Anything, invitationID).Return(nil, domain.ErrInvitationNotFound)

	err := service.AcceptInvitation(context.Background(), invitationID, userID)

	assert.ErrorIs(t, err, domain.ErrInvitationNotFound)
}

func TestGroupService_AcceptInvitation_WrongUser(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	invitationID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())
	otherUserID := uuid.Must(uuid.NewV4())

	invitation := &domain.Invitation{
		ID:     uuid.Must(uuid.FromString(invitationID)),
		UserID: otherUserID,
		Status: domain.StatusPending,
	}

	mockGroupRepo.On("GetInvitation", mock.Anything, invitationID).Return(invitation, nil)

	err := service.AcceptInvitation(context.Background(), invitationID, userID)

	assert.ErrorIs(t, err, domain.ErrInvitationNotFound)
}

func TestGroupService_LeaveGroup_Success(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, userID.String()).Return(domain.RoleMember, nil)
	mockGroupRepo.On("RemoveMember", mock.Anything, groupID, userID.String()).Return(nil)

	err := service.LeaveGroup(context.Background(), groupID, userID)

	assert.NoError(t, err)
}

func TestGroupService_LeaveGroup_Owner(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	userID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, userID.String()).Return(domain.RoleOwner, nil)

	err := service.LeaveGroup(context.Background(), groupID, userID)

	assert.ErrorIs(t, err, domain.ErrCannotRemoveGroupOwner)
}

func TestGroupService_RemoveMember_Success(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	ownerID := uuid.Must(uuid.NewV4())
	memberID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, ownerID.String()).Return(domain.RoleOwner, nil)
	mockGroupRepo.On("RemoveMember", mock.Anything, groupID, memberID.String()).Return(nil)

	err := service.RemoveMember(context.Background(), groupID, ownerID, memberID)

	assert.NoError(t, err)
}

func TestGroupService_RemoveMember_NotOwner(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	ownerID := uuid.Must(uuid.NewV4())
	memberID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, ownerID.String()).Return(domain.RoleMember, nil)

	err := service.RemoveMember(context.Background(), groupID, ownerID, memberID)

	assert.ErrorIs(t, err, domain.ErrNotGroupOwner)
}

func TestGroupService_TransferOwnership_Success(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	currentOwnerID := uuid.Must(uuid.NewV4())
	newOwnerID := uuid.Must(uuid.NewV4())

	group := &domain.Group{
		ID:      uuid.Must(uuid.FromString(groupID)),
		OwnerID: currentOwnerID,
	}

	mockGroupRepo.On("GetGroupByID", mock.Anything, groupID).Return(group, nil)
	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID, newOwnerID.String()).Return(domain.RoleMember, nil)
	mockGroupRepo.On("UpdateGroup", mock.Anything, group).Return(nil)
	mockGroupRepo.On("UpdateMemberRole", mock.Anything, groupID, currentOwnerID.String(), domain.RoleAdmin).Return(nil)
	mockGroupRepo.On("UpdateMemberRole", mock.Anything, groupID, newOwnerID.String(), domain.RoleOwner).Return(nil)

	err := service.TransferOwnership(context.Background(), groupID, currentOwnerID, newOwnerID)

	assert.NoError(t, err)
	assert.Equal(t, newOwnerID, group.OwnerID)
}

func TestGroupService_TransferOwnership_NotOwner(t *testing.T) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := NewGroupService(mockGroupRepo, mockUserRepo)

	groupID := uuid.Must(uuid.NewV4()).String()
	currentOwnerID := uuid.Must(uuid.NewV4())
	newOwnerID := uuid.Must(uuid.NewV4())

	group := &domain.Group{
		ID:      uuid.Must(uuid.FromString(groupID)),
		OwnerID: currentOwnerID,
	}

	mockGroupRepo.On("GetGroupByID", mock.Anything, groupID).Return(group, nil)

	err := service.TransferOwnership(context.Background(), groupID, uuid.Must(uuid.NewV4()), newOwnerID)

	assert.ErrorIs(t, err, domain.ErrNotGroupOwner)
}

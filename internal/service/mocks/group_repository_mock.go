package mocks

import (
	"context"

	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) CreateGroup(ctx context.Context, group *domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetGroupByID(ctx context.Context, id string) (*domain.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *MockGroupRepository) UpdateGroup(ctx context.Context, group *domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) DeleteGroup(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupRepository) GetUserGroups(ctx context.Context, userID string) ([]domain.Group, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Group), args.Error(1)
}

func (m *MockGroupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

func (m *MockGroupRepository) GetMembers(ctx context.Context, groupID string) ([]domain.GroupMember, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]domain.GroupMember), args.Error(1)
}

func (m *MockGroupRepository) GetMemberRole(ctx context.Context, groupID, userID string) (domain.Role, error) {
	args := m.Called(ctx, groupID, userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(domain.Role), args.Error(1)
}

func (m *MockGroupRepository) UpdateMemberRole(ctx context.Context, groupID, userID string, newRole domain.Role) error {
	args := m.Called(ctx, groupID, userID, newRole)
	return args.Error(0)
}

func (m *MockGroupRepository) CreateInvitation(ctx context.Context, inv *domain.Invitation) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockGroupRepository) GetInvitation(ctx context.Context, id string) (*domain.Invitation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invitation), args.Error(1)
}

func (m *MockGroupRepository) GetInvitationWithGroup(ctx context.Context, id string) (*domain.InvitationWithGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvitationWithGroup), args.Error(1)
}

func (m *MockGroupRepository) GetUserInvitationsWithGroup(ctx context.Context, userID string) ([]domain.InvitationWithGroup, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.InvitationWithGroup), args.Error(1)
}

func (m *MockGroupRepository) UpdateInvitation(ctx context.Context, inv *domain.Invitation) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockGroupRepository) GetGroupMembersCount(ctx context.Context, groupID string) (int, error) {
	args := m.Called(ctx, groupID)
	return args.Int(0), args.Error(1)
}

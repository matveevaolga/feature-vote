package repository

import (
	"context"

	"github.com/matveevaolga/feature-vote/internal/domain"
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *domain.Group) error
	GetGroupByID(ctx context.Context, id string) (*domain.Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]domain.Group, error)
	UpdateGroup(ctx context.Context, group *domain.Group) error
	DeleteGroup(ctx context.Context, id string) error

	AddMember(ctx context.Context, member *domain.GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	GetMembers(ctx context.Context, groupID string) ([]domain.GroupMember, error)
	GetMemberRole(ctx context.Context, groupID, userID string) (string, error)

	CreateInvitation(ctx context.Context, inv *domain.Invitation) error
	GetInvitation(ctx context.Context, id string) (*domain.Invitation, error)
	GetInvitationWithGroup(ctx context.Context, id string) (*domain.InvitationWithGroup, error)
	GetUserInvitationsWithGroup(ctx context.Context, userID string) ([]domain.Invitation, error)
	UpdateInvitation(ctx context.Context, inv *domain.Invitation) error
}

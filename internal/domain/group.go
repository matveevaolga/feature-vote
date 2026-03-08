package domain

import (
	"time"

	uuid "github.com/gofrs/uuid/v5"
)

type Role string

const (
	// RoleOwner has full control over the group. Can delete group, manage all members.
	RoleOwner Role = "owner"
	// RoleAdmin can manage members and create votings.
	RoleAdmin Role = "admin"
	// RoleMember can only participate in votings.
	RoleMember Role = "member"
)

type Status string

const (
	// StatusPending means the invitation is waiting for user's response.
	StatusPending Status = "pending"
	// StatusAccepted means the user has joined the group.
	StatusAccepted Status = "accepted"
	// StatusDeclined means the user has rejected the invitation.
	StatusDeclined Status = "declined"
)

type Group struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupMember struct {
	GroupID  uuid.UUID `json:"group_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     Role      `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Invitation struct {
	ID        uuid.UUID `json:"id"`
	GroupID   uuid.UUID `json:"group_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InvitationWithGroup struct {
	Invitation
	GroupName string
}

func NewGroup(name string, ownerID uuid.UUID) (*Group, *GroupMember) {
	groupID := uuid.Must(uuid.NewV4())

	group := &Group{
		ID:        groupID,
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
	}

	member := &GroupMember{
		GroupID:  groupID,
		UserID:   ownerID,
		Role:     RoleOwner,
		JoinedAt: time.Now(),
	}

	return group, member
}

func (g *Group) IsOwner(userID uuid.UUID) bool {
	return g.OwnerID == userID
}

func (g *Group) CanBeDeletedBy(userID uuid.UUID) bool {
	return g.IsOwner(userID)
}

func (inv *Invitation) Accept() error {
	if inv.Status != StatusPending {
		return ErrInvitationNotPending
	}
	inv.Status = StatusAccepted
	inv.UpdatedAt = time.Now()
	return nil
}

func (inv *Invitation) Decline() error {
	if inv.Status != StatusPending {
		return ErrInvitationNotPending
	}
	inv.Status = StatusDeclined
	inv.UpdatedAt = time.Now()
	return nil
}

func (inv *Invitation) IsForUser(userID uuid.UUID) bool {
	return inv.UserID == userID
}

func (inv *Invitation) IsPending() bool {
	return inv.Status == StatusPending
}

func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	}
	return false
}

func (r Role) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanCreateVoting() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanRemoveMember(targetRole Role) bool {
	if r == RoleOwner {
		return true
	}
	if r == RoleAdmin && targetRole == RoleMember {
		return true
	}
	return false
}

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusAccepted, StatusDeclined:
		return true
	}
	return false
}

func (s Status) IsResolved() bool {
	return s == StatusAccepted || s == StatusDeclined
}

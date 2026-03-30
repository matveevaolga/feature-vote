package domain

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestNewGroup(t *testing.T) {
	ownerID := uuid.Must(uuid.NewV4())
	groupName := "Test Group"

	group, member := NewGroup(groupName, ownerID)

	if group.Name != groupName {
		t.Errorf("expected group name %s, got %s", groupName, group.Name)
	}
	if group.OwnerID != ownerID {
		t.Errorf("expected owner %v, got %v", ownerID, group.OwnerID)
	}
	if member.Role != RoleOwner {
		t.Errorf("expected role owner, got %s", member.Role)
	}
	if member.UserID != ownerID {
		t.Errorf("expected member userID %v, got %v", ownerID, member.UserID)
	}
}

func TestGroup_IsOwner(t *testing.T) {
	ownerID := uuid.Must(uuid.NewV4())
	otherID := uuid.Must(uuid.NewV4())

	group, _ := NewGroup("Test", ownerID)

	if !group.IsOwner(ownerID) {
		t.Error("expected true for owner")
	}
	if group.IsOwner(otherID) {
		t.Error("expected false for non-owner")
	}
}

func TestInvitation_Accept(t *testing.T) {
	inv := &Invitation{
		ID:        uuid.Must(uuid.NewV4()),
		GroupID:   uuid.Must(uuid.NewV4()),
		UserID:    uuid.Must(uuid.NewV4()),
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := inv.Accept()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if inv.Status != StatusAccepted {
		t.Errorf("expected status accepted, got %s", inv.Status)
	}
}

func TestInvitation_Accept_InvalidStatus(t *testing.T) {
	inv := &Invitation{
		Status: StatusAccepted,
	}
	err := inv.Accept()
	if err != ErrInvitationNotPending {
		t.Errorf("expected ErrInvitationNotPending, got %v", err)
	}
}

func TestRole_CanManageMembers(t *testing.T) {
	tests := []struct {
		role     Role
		expected bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleMember, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.CanManageMembers(); got != tt.expected {
				t.Errorf("CanManageMembers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

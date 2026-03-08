package dto

import (
	"time"
)

type CreateGroupRequest struct {
	Name string `json:"name" validate:"required,min=5,max=50"`
}

type UpdateGroupRequest struct {
	Name string `json:"name" validate:"required,min=5,max=50"`
}

type GroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupMemberResponse struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type InviteMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type InvitationResponse struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateRoleRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Role   string `json:"role" validate:"required,oneof=admin member"`
}

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"new_owner_id" validate:"required,uuid"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUsername   = errors.New("invalid username")

	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupAlreadyExists = errors.New("group already exists")
	ErrNotGroupOwner      = errors.New("user is not group owner")
	ErrNotGroupMember     = errors.New("user is not group member")
	ErrAlreadyGroupMember = errors.New("user is already group member")

	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationNotPending  = errors.New("invitation is not pending")
	ErrInvitationAlreadySent = errors.New("invitation already sent")
)

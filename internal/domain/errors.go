package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid email or password")

	ErrGroupNotFound          = errors.New("group not found")
	ErrGroupAlreadyExists     = errors.New("group already exists")
	ErrNotGroupOwner          = errors.New("user is not group owner")
	ErrNotGroupMember         = errors.New("user is not group member")
	ErrAlreadyGroupMember     = errors.New("user is already group member")
	ErrInvalidGroupName       = errors.New("group name must be between 5 and 50 characters")
	ErrCannotRemoveGroupOwner = errors.New("cannot leave the group as owner, delete the group instead")

	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationNotPending  = errors.New("invitation is not pending")
	ErrInvitationAlreadySent = errors.New("invitation already sent")

	ErrVotingNotFound     = errors.New("voting not found")
	ErrVotingNotActive    = errors.New("voting is not active")
	ErrAlreadyVoted       = errors.New("user already voted in this voting")
	ErrVoteNotFound       = errors.New("vote not found")
	ErrVotingServiceBusy  = errors.New("the voting service is busy, try later")
	ErrFailedToStopVoting = errors.New("failed to stop voting, try again")
)

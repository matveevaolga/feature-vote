package main

import "time"

type SeedData struct {
	Users       []UserSeed       `json:"users"`
	Groups      []GroupSeed      `json:"groups"`
	Members     []MemberSeed     `json:"members"`
	Invitations []InvitationSeed `json:"invitations"`
	Votings     []VotingSeed     `json:"votings"`
	Votes       []VoteSeed       `json:"votes"`
}

type UserSeed struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupSeed struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type MemberSeed struct {
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type InvitationSeed struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VotingSeed struct {
	ID          string     `json:"id"`
	GroupID     string     `json:"group_id"`
	CreatedBy   string     `json:"created_by"`
	FeatureName string     `json:"feature_name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Duration    int64      `json:"duration"`
	StartedAt   time.Time  `json:"started_at"`
	EndsAt      time.Time  `json:"ends_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Result      *bool      `json:"result"`
}

type VoteSeed struct {
	VotingID  string    `json:"voting_id"`
	UserID    string    `json:"user_id"`
	Vote      string    `json:"vote"`
	CreatedAt time.Time `json:"created_at"`
}

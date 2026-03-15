package dto

import (
	"time"
)

type CreateVotingRequest struct {
	GroupID     string `json:"group_id" validate:"required,uuid"`
	FeatureName string `json:"feature_name" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"max=1000"`
	DurationSec int    `json:"duration_sec" validate:"required,min=10,max=3600"`
}

type CreateVotingResponse struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	FeatureName string    `json:"feature_name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	EndsAt      time.Time `json:"ends_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type CastVoteRequest struct {
	Vote string `json:"vote" validate:"required,oneof=yes no"`
}

type VotingResultResponse struct {
	VotingID     string  `json:"voting_id"`
	TotalVotes   int     `json:"total_votes"`
	YesVotes     int     `json:"yes_votes"`
	NoVotes      int     `json:"no_votes"`
	YesPercent   float64 `json:"yes_percent"`
	NoPercent    float64 `json:"no_percent"`
	Participated int     `json:"participated"`
	TotalMembers int     `json:"total_members"`
	IsCompleted  bool    `json:"is_completed"`
	Result       *bool   `json:"result,omitempty"`
}

type VotingStatusResponse struct {
	ID          string     `json:"id"`
	GroupID     string     `json:"group_id"`
	FeatureName string     `json:"feature_name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	EndsAt      time.Time  `json:"ends_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      *bool      `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type VotingListResponse struct {
	Votings []VotingStatusResponse `json:"votings"`
	Total   int                    `json:"total"`
	Page    int                    `json:"page"`
	PerPage int                    `json:"per_page"`
}

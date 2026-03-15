package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type VoteType string

const (
	VoteYes VoteType = "yes"
	VoteNo  VoteType = "no"
)

type VotingStatus string

const (
	VotingStatusActive    VotingStatus = "active"
	VotingStatusCompleted VotingStatus = "completed"
	VotingStatusCancelled VotingStatus = "cancelled"
)

type Voting struct {
	ID          uuid.UUID
	GroupID     uuid.UUID
	CreatedBy   uuid.UUID
	FeatureName string
	Description string
	Status      VotingStatus
	Duration    time.Duration
	StartedAt   time.Time
	EndsAt      time.Time
	CompletedAt *time.Time
	Result      *bool
}

type Vote struct {
	VotingID  uuid.UUID
	UserID    uuid.UUID
	Vote      VoteType
	CreatedAt time.Time
}

type VotingResult struct {
	VotingID     uuid.UUID
	TotalVotes   int
	YesVotes     int
	NoVotes      int
	YesPercent   float64
	NoPercent    float64
	Participated int
	TotalMembers int
	IsCompleted  bool
}

func NewVoting(groupID, createdBy uuid.UUID, featureName, description string, duration time.Duration) *Voting {
	now := time.Now()
	return &Voting{
		ID:          uuid.Must(uuid.NewV4()),
		GroupID:     groupID,
		CreatedBy:   createdBy,
		FeatureName: featureName,
		Description: description,
		Status:      VotingStatusActive,
		Duration:    duration,
		StartedAt:   now,
		EndsAt:      now.Add(duration),
	}
}

func (v *Voting) IsActive() bool {
	return v.Status == VotingStatusActive && time.Now().Before(v.EndsAt)
}

func (v *Voting) HasEnded() bool {
	return time.Now().After(v.EndsAt) || v.Status != VotingStatusActive
}

func (v *Voting) Complete(result bool) {
	v.Status = VotingStatusCompleted
	now := time.Now()
	v.CompletedAt = &now
	v.Result = &result
}

func (v *Voting) Cancel() {
	v.Status = VotingStatusCancelled
	now := time.Now()
	v.CompletedAt = &now
}

func (v *Vote) IsValid() bool {
	return v.Vote == VoteYes || v.Vote == VoteNo
}

func (vr *VotingResult) Calculate() {
	if vr.TotalVotes > 0 {
		vr.YesPercent = float64(vr.YesVotes) / float64(vr.TotalVotes) * 100
		vr.NoPercent = float64(vr.NoVotes) / float64(vr.TotalVotes) * 100
	}
}

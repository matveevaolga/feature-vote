package domain

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestNewVoting(t *testing.T) {
	groupID := uuid.Must(uuid.NewV4())
	createdBy := uuid.Must(uuid.NewV4())
	duration := 5 * time.Minute

	voting := NewVoting(groupID, createdBy, "Add Dark Mode", "Enable dark theme", duration)

	if voting.GroupID != groupID {
		t.Errorf("expected groupID %v, got %v", groupID, voting.GroupID)
	}
	if voting.CreatedBy != createdBy {
		t.Errorf("expected createdBy %v, got %v", createdBy, voting.CreatedBy)
	}
	if voting.FeatureName != "Add Dark Mode" {
		t.Errorf("expected feature name 'Add Dark Mode', got '%s'", voting.FeatureName)
	}
	if voting.Status != VotingStatusActive {
		t.Errorf("expected status active, got %s", voting.Status)
	}
	if voting.Duration != duration {
		t.Errorf("expected duration %v, got %v", duration, voting.Duration)
	}
}

func TestVoting_IsActive(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		status   VotingStatus
		endsAt   time.Time
		expected bool
	}{
		{"active and not ended", VotingStatusActive, now.Add(5 * time.Minute), true},
		{"active but ended", VotingStatusActive, now.Add(-5 * time.Minute), false},
		{"completed", VotingStatusCompleted, now.Add(5 * time.Minute), false},
		{"cancelled", VotingStatusCancelled, now.Add(5 * time.Minute), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voting := &Voting{
				Status: tt.status,
				EndsAt: tt.endsAt,
			}
			if got := voting.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVoting_Complete(t *testing.T) {
	voting := NewVoting(
		uuid.Must(uuid.NewV4()),
		uuid.Must(uuid.NewV4()),
		"Feature",
		"Description",
		5*time.Minute,
	)

	result := true
	voting.Complete(result)

	if voting.Status != VotingStatusCompleted {
		t.Errorf("expected status completed, got %s", voting.Status)
	}
	if voting.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if *voting.Result != result {
		t.Errorf("expected result %v, got %v", result, *voting.Result)
	}
}

func TestVote_IsValid(t *testing.T) {
	tests := []struct {
		voteType VoteType
		expected bool
	}{
		{VoteYes, true},
		{VoteNo, true},
		{VoteType("maybe"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.voteType), func(t *testing.T) {
			vote := &Vote{Vote: tt.voteType}
			if got := vote.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/domain/repository"
)

type ActiveVoting struct {
	Voting   *domain.Voting
	VoteChan chan *domain.Vote
	StopChan chan struct{}
	mu       sync.RWMutex
	votes    map[uuid.UUID]domain.VoteType
	ctx      context.Context
	cancel   context.CancelFunc
}

type VotingService struct {
	votingRepo repository.VotingRepository
	groupRepo  repository.GroupRepository
	userRepo   repository.UserRepository

	activeVotings sync.Map
	mu            sync.RWMutex
}

func NewVotingService(votingRepo repository.VotingRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository) *VotingService {
	return &VotingService{
		votingRepo:    votingRepo,
		groupRepo:     groupRepo,
		userRepo:      userRepo,
		activeVotings: sync.Map{},
	}
}

func (s *VotingService) CreateVoting(ctx context.Context, groupID string,
	createdBy uuid.UUID, featureName, description string, duration time.Duration) (*domain.Voting, error) {
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, createdBy.String())
	if err != nil {
		return nil, err
	}
	if !role.CanCreateVoting() {
		return nil, domain.ErrNotGroupOwner
	}
	voting := domain.NewVoting(
		uuid.Must(uuid.FromString(groupID)), createdBy,
		featureName, description, duration)
	if err := s.votingRepo.CreateVoting(ctx, voting); err != nil {
		return nil, err
	}
	s.startVoting(voting)
	return voting, nil
}

func (s *VotingService) startVoting(voting *domain.Voting) {
	ctx, cancel := context.WithDeadline(context.Background(), voting.EndsAt)
	active := &ActiveVoting{
		Voting:   voting,
		VoteChan: make(chan *domain.Vote, 100),
		StopChan: make(chan struct{}),
		votes:    make(map[uuid.UUID]domain.VoteType),
		ctx:      ctx,
		cancel:   cancel,
	}
	s.activeVotings.Store(voting.ID.String(), active)
	go s.runVoting(active)
}

func (s *VotingService) runVoting(active *ActiveVoting) {
	defer func() {
		s.activeVotings.Delete(active.Voting.ID.String())
		close(active.VoteChan)
		close(active.StopChan)
		active.cancel()
	}()

	for {
		select {
		case <-active.ctx.Done():
			s.finalizeVoting(active)
			return
		case <-active.StopChan:
			s.cancelVoting(active)
			return
		case vote := <-active.VoteChan:
			s.processVote(active, vote)
		}
	}
}

func (s *VotingService) processVote(active *ActiveVoting, vote *domain.Vote) {
	select {
	case <-active.ctx.Done():
		return
	default:
	}
	active.mu.Lock()
	defer active.mu.Unlock()
	if _, hasVoted := active.votes[vote.UserID]; hasVoted {
		return
	}
	active.votes[vote.UserID] = vote.Vote
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.votingRepo.CastVote(ctx, vote); err != nil {
			slog.Error("failed to cast vote to DB",
				"voting_id", vote.VotingID,
				"user_id", vote.UserID,
				"error", err,
			)
		}
	}()
}

func (s *VotingService) finalizeVoting(active *ActiveVoting) {
	active.mu.RLock()
	defer active.mu.RUnlock()
	yesCnt, noCnt := 0, 0
	for _, voteType := range active.votes {
		if voteType == domain.VoteYes {
			yesCnt++
		} else {
			noCnt++
		}
	}
	result := yesCnt > noCnt
	active.Voting.Complete(result)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.votingRepo.UpdateVoting(ctx, active.Voting); err != nil {
		slog.Error("failed to update voting result in DB",
			"voting_id", active.Voting.ID,
			"result", active.Voting.Result,
			"error", err,
		)
	}
}

func (s *VotingService) cancelVoting(active *ActiveVoting) {
	active.Voting.Cancel()
	active.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.votingRepo.UpdateVoting(ctx, active.Voting); err != nil {
		slog.Error("failed to update cancelled voting in DB",
			"voting_id", active.Voting.ID,
			"error", err,
		)
	}
}

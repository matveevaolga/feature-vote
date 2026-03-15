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
		slog.Info("duplicate vote ignored", "voting_id", vote.VotingID, "user_id", vote.UserID)
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

func (s *VotingService) CastVote(ctx context.Context, votingID string, userID uuid.UUID, voteType domain.VoteType) error {
	value, ok := s.activeVotings.Load(votingID)
	if !ok {
		voting, err := s.votingRepo.GetVotingByID(ctx, votingID)
		if err != nil {
			return err
		}
		if voting.HasEnded() {
			return domain.ErrVotingNotActive
		}
		return domain.ErrVotingNotFound
	}
	active := value.(*ActiveVoting)
	if _, err := s.groupRepo.GetMemberRole(ctx, active.Voting.GroupID.String(), userID.String()); err != nil {
		return domain.ErrNotGroupMember
	}
	select {
	case <-active.ctx.Done():
		return domain.ErrVotingNotActive
	default:
	}
	vote := &domain.Vote{
		VotingID:  active.Voting.ID,
		UserID:    userID,
		Vote:      voteType,
		CreatedAt: time.Now(),
	}
	select {
	case active.VoteChan <- vote:
		slog.Debug("vote cast", "voting_id", votingID, "user_id", userID)
		return nil
	case <-time.After(5 * time.Second):
		return domain.ErrVotingServiceBusy
	case <-active.ctx.Done():
		return domain.ErrVotingNotActive
	}
}

func (s *VotingService) StopVoting(ctx context.Context, votingID string, userID uuid.UUID) error {
	value, ok := s.activeVotings.Load(votingID)
	if !ok {
		return domain.ErrVotingNotFound
	}
	active := value.(*ActiveVoting)
	if active.Voting.CreatedBy != userID {
		role, err := s.groupRepo.GetMemberRole(ctx, active.Voting.GroupID.String(), userID.String())
		if err != nil {
			return err
		}
		if !role.CanCreateVoting() {
			return domain.ErrNotGroupOwner
		}
	}
	select {
	case active.StopChan <- struct{}{}:
		slog.Info("voting stopped", "voting_id", votingID, "user_id", userID)
		return nil
	case <-time.After(5 * time.Second):
		return domain.ErrFailedToStopVoting
	}
}

func (s *VotingService) GetVotingResult(ctx context.Context, votingID string, userID uuid.UUID) (*domain.VotingResult, error) {
	value, ok := s.activeVotings.Load(votingID)
	if ok {
		active := value.(*ActiveVoting)
		if _, err := s.groupRepo.GetMemberRole(ctx, active.Voting.GroupID.String(), userID.String()); err != nil {
			return nil, domain.ErrNotGroupMember
		}
		active.mu.RLock()
		defer active.mu.RUnlock()
		yesCnt, noCnt := 0, 0
		for _, vt := range active.votes {
			if vt == domain.VoteYes {
				yesCnt++
			} else {
				noCnt++
			}
		}
		totalMembers, err := s.groupRepo.GetGroupMembersCount(ctx, active.Voting.GroupID.String())
		if err != nil {
			return nil, err
		}
		voting, err := s.GetVotingByID(ctx, votingID, userID)
		if err != nil {
			return nil, err
		}
		result := &domain.VotingResult{
			VotingID:     active.Voting.ID,
			TotalVotes:   len(active.votes),
			YesVotes:     yesCnt,
			NoVotes:      noCnt,
			Participated: len(active.votes),
			TotalMembers: totalMembers,
			IsCompleted: voting.Status == domain.VotingStatusCompleted ||
				voting.Status == domain.VotingStatusCancelled,
		}
		result.Calculate()
		return result, nil
	}
	voting, err := s.votingRepo.GetVotingByID(ctx, votingID)
	if err != nil {
		return nil, err
	}
	if _, err := s.groupRepo.GetMemberRole(ctx, voting.GroupID.String(), userID.String()); err != nil {
		return nil, domain.ErrNotGroupMember
	}
	yesCnt, noCnt, err := s.votingRepo.CountVotesByType(ctx, votingID)
	if err != nil {
		return nil, err
	}
	totalVotes, err := s.votingRepo.CountVotes(ctx, votingID)
	if err != nil {
		return nil, err
	}
	totalMembers, err := s.groupRepo.GetGroupMembersCount(ctx, voting.GroupID.String())
	if err != nil {
		return nil, err
	}
	result := &domain.VotingResult{
		VotingID:     voting.ID,
		TotalVotes:   totalVotes,
		YesVotes:     yesCnt,
		NoVotes:      noCnt,
		Participated: totalVotes,
		TotalMembers: totalMembers,
	}
	result.Calculate()
	return result, nil
}

func (s *VotingService) GetVotingByID(ctx context.Context, votingID string, userID uuid.UUID) (*domain.Voting, error) {
	value, ok := s.activeVotings.Load(votingID)
	if ok {
		active := value.(*ActiveVoting)
		if _, err := s.groupRepo.GetMemberRole(ctx, active.Voting.GroupID.String(), userID.String()); err != nil {
			return nil, domain.ErrNotGroupMember
		}
		return active.Voting, nil
	}
	voting, err := s.votingRepo.GetVotingByID(ctx, votingID)
	if err != nil {
		return nil, err
	}
	if _, err := s.groupRepo.GetMemberRole(ctx, voting.GroupID.String(), userID.String()); err != nil {
		return nil, domain.ErrNotGroupMember
	}
	return voting, nil
}

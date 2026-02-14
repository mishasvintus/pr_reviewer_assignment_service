package service

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
)

// ReviewerAssigner handles reviewer selection logic.
type ReviewerAssigner struct{}

// NewReviewerAssigner creates a new reviewer assigner.
func NewReviewerAssigner() *ReviewerAssigner {
	return &ReviewerAssigner{}
}

// SelectReviewers selects up to 2 reviewers from active teammates.
// Uses cryptographically secure random selection.
func (a *ReviewerAssigner) SelectReviewers(teammates []domain.User) ([]string, error) {
	if len(teammates) == 0 {
		return []string{}, nil
	}

	if len(teammates) <= 2 {
		reviewers := make([]string, len(teammates))
		for i, user := range teammates {
			reviewers[i] = user.UserID
		}
		return reviewers, nil
	}

	selected := make(map[int]bool)
	reviewers := make([]string, 0, 2)

	for len(reviewers) < 2 {
		idx, err := secureRandInt(len(teammates))
		if err != nil {
			return nil, fmt.Errorf("failed to generate random index: %w", err)
		}

		if !selected[idx] {
			selected[idx] = true
			reviewers = append(reviewers, teammates[idx].UserID)
		}
	}

	return reviewers, nil
}

// SelectReassignReviewers selects up to 2 new reviewers, excluding author and currently assigned reviewers.
func (a *ReviewerAssigner) SelectReassignReviewers(teammates []domain.User, authorID string, assignedReviewers []string) ([]string, error) {
	excludeIDs := make(map[string]struct{})
	excludeIDs[authorID] = struct{}{}
	for _, id := range assignedReviewers {
		excludeIDs[id] = struct{}{}
	}

	candidates := make([]domain.User, 0)
	for _, user := range teammates {
		if _, excluded := excludeIDs[user.UserID]; !excluded {
			candidates = append(candidates, user)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates available for reassignment")
	}

	return a.SelectReviewers(candidates)
}

// secureRandInt returns a cryptographically secure random integer in [0, max).
func secureRandInt(max int) (int, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(nBig.Int64()), nil
}

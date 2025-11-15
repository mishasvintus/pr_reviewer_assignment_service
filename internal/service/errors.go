package service

import "errors"

var (
	ErrTeamExists          = errors.New("team already exists")
	ErrTeamNotFound        = errors.New("team not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrPRAuthorNotFound    = errors.New("author not found")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRMerged            = errors.New("cannot reassign merged pull request")
	ErrReviewerNotAssigned = errors.New("user is not assigned to this pull request")
	ErrNoCandidate         = errors.New("no candidates available for reassignment")
	ErrInactiveReviewer    = errors.New("reviewer is not active")
)

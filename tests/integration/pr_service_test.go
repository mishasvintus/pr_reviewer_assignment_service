package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/pr"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/team"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
	"github.com/mishasvintus/avito_backend_internship/tests"
)

func TestPRService_CreatePR(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	// create team and author
	teamName := "team1"
	authorID := "author1"
	require.NoError(t, team.Create(db, teamName))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   authorID,
		Username: "author",
		TeamName: teamName,
		IsActive: true,
	}))

	assigner := service.NewReviewerAssigner()
	prService := service.NewPRService(db, assigner)

	t.Run("success - creates PR with reviewers", func(t *testing.T) {
		// Create teammates
		reviewer1ID := "reviewer1"
		reviewer2ID := "reviewer2"
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   reviewer1ID,
			Username: "reviewer1",
			TeamName: teamName,
			IsActive: true,
		}))
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   reviewer2ID,
			Username: "reviewer2",
			TeamName: teamName,
			IsActive: true,
		}))

		prID := "pr1"
		prName := "Test PR"

		createdPR, err := prService.CreatePR(prID, prName, authorID)
		require.NoError(t, err)
		assert.Equal(t, prID, createdPR.PullRequestID)
		assert.Equal(t, prName, createdPR.PullRequestName)
		assert.Equal(t, authorID, createdPR.AuthorID)
		assert.Equal(t, domain.StatusOpen, createdPR.Status)
		assert.GreaterOrEqual(t, len(createdPR.AssignedReviewers), 1)
		assert.LessOrEqual(t, len(createdPR.AssignedReviewers), 2)
	})

	t.Run("error - author not found", func(t *testing.T) {
		_, err := prService.CreatePR("pr2", "Test PR", "nonexistent")
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrPRAuthorNotFound))
	})

	t.Run("error - PR already exists", func(t *testing.T) {
		prID := "pr3"
		prName := "Test PR"

		// Create teammates for first PR
		reviewer1ID := "reviewer3"
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   reviewer1ID,
			Username: "reviewer3",
			TeamName: teamName,
			IsActive: true,
		}))

		// Create PR first time
		_, err := prService.CreatePR(prID, prName, authorID)
		require.NoError(t, err)

		// Try to create again
		_, err = prService.CreatePR(prID, prName, authorID)
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrPRExists))
	})
}

func TestPRService_MergePR(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	// Setup: create team, author, and PR
	teamName := "team1"
	authorID := "author1"
	prID := "pr1"

	require.NoError(t, team.Create(db, teamName))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   authorID,
		Username: "author",
		TeamName: teamName,
		IsActive: true,
	}))

	assigner := service.NewReviewerAssigner()
	prService := service.NewPRService(db, assigner)

	t.Run("success - merges PR", func(t *testing.T) {
		// Create PR
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID,
			PullRequestName: "Test PR",
			AuthorID:        authorID,
			Status:          domain.StatusOpen,
		}))

		mergedPR, err := prService.MergePR(prID)
		require.NoError(t, err)
		assert.Equal(t, prID, mergedPR.PullRequestID)
		assert.Equal(t, domain.StatusMerged, mergedPR.Status)
		assert.NotNil(t, mergedPR.MergedAt)
	})

	t.Run("success - idempotent merge", func(t *testing.T) {
		// PR already merged, should return without error
		mergedPR, err := prService.MergePR(prID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusMerged, mergedPR.Status)
	})

	t.Run("error - PR not found", func(t *testing.T) {
		_, err := prService.MergePR("nonexistent")
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrPRNotFound))
	})
}

func TestPRService_ReassignPR(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	// Setup: create team, users
	teamName := "team1"
	authorID := "author1"
	oldReviewerID := "reviewer1"
	newReviewerID := "reviewer2"

	require.NoError(t, team.Create(db, teamName))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   authorID,
		Username: "author",
		TeamName: teamName,
		IsActive: true,
	}))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   oldReviewerID,
		Username: "old_reviewer",
		TeamName: teamName,
		IsActive: true,
	}))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   newReviewerID,
		Username: "new_reviewer",
		TeamName: teamName,
		IsActive: true,
	}))

	assigner := service.NewReviewerAssigner()
	prService := service.NewPRService(db, assigner)

	t.Run("success - reassigns reviewer", func(t *testing.T) {
		prID := "pr1"
		// Create PR with old reviewer
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID,
			PullRequestName: "Test PR",
			AuthorID:        authorID,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, oldReviewerID))

		updatedPR, replacedBy, err := prService.ReassignPR(prID, oldReviewerID)
		require.NoError(t, err)
		assert.Equal(t, prID, updatedPR.PullRequestID)
		assert.Equal(t, newReviewerID, replacedBy)
		assert.Contains(t, updatedPR.AssignedReviewers, newReviewerID)
		assert.NotContains(t, updatedPR.AssignedReviewers, oldReviewerID)
	})

	t.Run("error - PR not found", func(t *testing.T) {
		_, _, err := prService.ReassignPR("nonexistent", oldReviewerID)
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrPRNotFound))
	})

	t.Run("error - PR already merged", func(t *testing.T) {
		prID := "pr2"
		// Create and merge PR
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID,
			PullRequestName: "Merged PR",
			AuthorID:        authorID,
			Status:          domain.StatusMerged,
		}))

		_, _, err := prService.ReassignPR(prID, oldReviewerID)
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrPRMerged))
	})

	t.Run("error - reviewer not assigned", func(t *testing.T) {
		prID := "pr3"
		// Create another reviewer in the same team (assigned to PR)
		assignedReviewerID := "reviewer3"
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   assignedReviewerID,
			Username: "reviewer3",
			TeamName: teamName,
			IsActive: true,
		}))

		// Create unassigned reviewer in the same team (will be candidate)
		unassignedReviewerID := "reviewer4"
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   unassignedReviewerID,
			Username: "reviewer4",
			TeamName: teamName,
			IsActive: true,
		}))

		// Create PR with assigned reviewer
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID,
			PullRequestName: "Test PR",
			AuthorID:        authorID,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, assignedReviewerID))

		// Try to reassign reviewer that is not assigned (but exists in team)
		_, _, err := prService.ReassignPR(prID, unassignedReviewerID)
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrReviewerNotAssigned))
	})
}

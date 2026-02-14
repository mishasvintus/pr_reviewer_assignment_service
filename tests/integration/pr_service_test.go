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
		assert.GreaterOrEqual(t, len(createdPR.AssignedReviewersIDs), 1)
		assert.LessOrEqual(t, len(createdPR.AssignedReviewersIDs), 2)
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
			TeamName:        teamName,
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

func TestPRService_ReplenishReviewers(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	prService := service.NewPRService(db, service.NewReviewerAssigner())

	t.Run("no error when PR not found", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		err = prService.ReplenishReviewers(tx, "nonexistent_pr")
		require.NoError(t, err)
	})

	t.Run("no op when PR is merged", func(t *testing.T) {
		teamName := "team_repl"
		authorID := "author_repl"
		require.NoError(t, team.Create(db, teamName))
		require.NoError(t, user.Create(db, &domain.User{UserID: authorID, Username: "author_repl", TeamName: teamName, IsActive: true}))
		prID := "pr_merged_repl"
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID: prID, PullRequestName: "Merged", AuthorID: authorID, TeamName: teamName, Status: domain.StatusMerged,
		}))
		tx, err := db.Begin()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		err = prService.ReplenishReviewers(tx, prID)
		require.NoError(t, err)
	})

	t.Run("no op when PR already has 2 reviewers", func(t *testing.T) {
		teamName := "team_repl2"
		authorID := "author_repl2"
		r1, r2 := "r1_repl2", "r2_repl2"
		require.NoError(t, team.Create(db, teamName))
		require.NoError(t, user.Create(db, &domain.User{UserID: authorID, Username: "a", TeamName: teamName, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r1, Username: "r1", TeamName: teamName, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r2, Username: "r2", TeamName: teamName, IsActive: true}))
		prID := "pr_full_repl"
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID: prID, PullRequestName: "Full", AuthorID: authorID, TeamName: teamName, Status: domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, r1))
		require.NoError(t, pr.InsertReviewer(db, prID, r2))
		tx, err := db.Begin()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		err = prService.ReplenishReviewers(tx, prID)
		require.NoError(t, err)
	})

	t.Run("no op when no candidates to replenish", func(t *testing.T) {
		// PR has 1 reviewer; team has only author + that reviewer -> no candidates
		teamName := "team_repl_nocand"
		authorID := "author_repl_nocand"
		r1 := "r1_repl_nocand"
		require.NoError(t, team.Create(db, teamName))
		require.NoError(t, user.Create(db, &domain.User{UserID: authorID, Username: "a", TeamName: teamName, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r1, Username: "r1", TeamName: teamName, IsActive: true}))
		prID := "pr_nocand_repl"
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID: prID, PullRequestName: "NoCand", AuthorID: authorID, TeamName: teamName, Status: domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, r1))
		tx, err := db.Begin()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		err = prService.ReplenishReviewers(tx, prID)
		require.NoError(t, err)
	})

	t.Run("replenishes when PR has less than 2 reviewers", func(t *testing.T) {
		teamName := "team_repl_add"
		authorID := "author_repl_add"
		r1, r2 := "r1_repl_add", "r2_repl_add"
		require.NoError(t, team.Create(db, teamName))
		require.NoError(t, user.Create(db, &domain.User{UserID: authorID, Username: "a", TeamName: teamName, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r1, Username: "r1", TeamName: teamName, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r2, Username: "r2", TeamName: teamName, IsActive: true}))
		prID := "pr_replenish_me"
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID: prID, PullRequestName: "Repl", AuthorID: authorID, TeamName: teamName, Status: domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, r1))
		tx, err := db.Begin()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		err = prService.ReplenishReviewers(tx, prID)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())
		updated, err := pr.Get(db, prID)
		require.NoError(t, err)
		assert.Len(t, updated.AssignedReviewersIDs, 2)
		assert.Contains(t, updated.AssignedReviewersIDs, r1)
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
			TeamName:        teamName,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, oldReviewerID))

		updatedPR, replacedBy, err := prService.ReassignPR(prID, oldReviewerID)
		require.NoError(t, err)
		assert.Equal(t, prID, updatedPR.PullRequestID)
		assert.Equal(t, newReviewerID, replacedBy)
		assert.Contains(t, updatedPR.AssignedReviewersIDs, newReviewerID)
		assert.NotContains(t, updatedPR.AssignedReviewersIDs, oldReviewerID)
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
			TeamName:        teamName,
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
			TeamName:        teamName,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, assignedReviewerID))

		// Try to reassign reviewer that is not assigned (but exists in team)
		_, _, err := prService.ReassignPR(prID, unassignedReviewerID)
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrReviewerNotAssigned))
	})

	t.Run("error - no candidate for reassignment", func(t *testing.T) {
		// Team of exactly 3: author + 2 reviewers. PR has both reviewers assigned.
		// Reassigning one leaves no candidate (author and other reviewer excluded).
		teamNameNC := "team_no_candidate"
		authorIDNC := "author_nc"
		r1ID := "reviewer_nc_1"
		r2ID := "reviewer_nc_2"
		require.NoError(t, team.Create(db, teamNameNC))
		require.NoError(t, user.Create(db, &domain.User{UserID: authorIDNC, Username: "author_nc", TeamName: teamNameNC, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r1ID, Username: "r1_nc", TeamName: teamNameNC, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: r2ID, Username: "r2_nc", TeamName: teamNameNC, IsActive: true}))

		prID := "pr_no_candidate"
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID,
			PullRequestName: "PR",
			AuthorID:        authorIDNC,
			TeamName:        teamNameNC,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.InsertReviewer(db, prID, r1ID))
		require.NoError(t, pr.InsertReviewer(db, prID, r2ID))

		_, _, err := prService.ReassignPR(prID, r1ID)
		assert.Error(t, err)
		assert.True(t, assert.ErrorIs(t, err, service.ErrNoCandidate))
	})
}

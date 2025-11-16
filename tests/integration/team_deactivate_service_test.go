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

func TestTeamService_DeactivateTeam(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer db.Close()
	defer tests.CleanupTestDB(db)

	teamService := service.NewTeamService(db)

	t.Run("success - deactivates team without PRs", func(t *testing.T) {
		teamName := "team_no_prs"
		userID1 := "user_no_prs_1"
		userID2 := "user_no_prs_2"

		require.NoError(t, team.Create(db, teamName))
		require.NoError(t, user.Create(db, &domain.User{UserID: userID1, Username: "User1", TeamName: teamName, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: userID2, Username: "User2", TeamName: teamName, IsActive: true}))

		err := teamService.DeactivateTeam(teamName)
		require.NoError(t, err)

		// Verify users are inactive
		u1, err := user.Get(db, userID1)
		require.NoError(t, err)
		assert.False(t, u1.IsActive)

		u2, err := user.Get(db, userID2)
		require.NoError(t, err)
		assert.False(t, u2.IsActive)
	})

	t.Run("error - team not found", func(t *testing.T) {
		err := teamService.DeactivateTeam("nonexistent_team")
		assert.ErrorIs(t, err, service.ErrTeamNotFound)
	})

	t.Run("success - deactivates team and reassigns PRs to author team", func(t *testing.T) {
		// Setup: 2 teams
		teamToDeactivate := "team_deact_with_prs"
		authorTeam := "author_team_with_prs"
		authorID := "author_with_prs"
		reviewerID := "reviewer_with_prs"
		teammateID := "teammate_with_prs"

		require.NoError(t, team.Create(db, teamToDeactivate))
		require.NoError(t, team.Create(db, authorTeam))
		require.NoError(t, user.Create(db, &domain.User{UserID: reviewerID, Username: "Reviewer", TeamName: teamToDeactivate, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: authorID, Username: "Author", TeamName: authorTeam, IsActive: true}))
		require.NoError(t, user.Create(db, &domain.User{UserID: teammateID, Username: "Teammate", TeamName: authorTeam, IsActive: true}))

		prID := "pr-deact-1"
		require.NoError(t, pr.Create(db, &domain.PullRequest{PullRequestID: prID, PullRequestName: "PR 1", AuthorID: authorID, Status: domain.StatusOpen}))
		require.NoError(t, pr.InsertReviewer(db, prID, reviewerID))

		err := teamService.DeactivateTeam(teamToDeactivate)
		require.NoError(t, err)

		// Verify reviewer is inactive
		uRev, err := user.Get(db, reviewerID)
		require.NoError(t, err)
		assert.False(t, uRev.IsActive)

		// Verify PR is updated
		pullRequest, err := pr.Get(db, prID)
		require.NoError(t, err)
		assert.Len(t, pullRequest.AssignedReviewers, 1)
		assert.NotEqual(t, reviewerID, pullRequest.AssignedReviewers[0])
		assert.Equal(t, teammateID, pullRequest.AssignedReviewers[0])
	})
}

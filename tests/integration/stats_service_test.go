package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/pr"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/stats"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/team"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
	"github.com/mishasvintus/avito_backend_internship/tests"
)

func TestStatsService_GetStatistics(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	statsService := service.NewStatsService(db)

	t.Run("success - empty statistics", func(t *testing.T) {
		stats, err := statsService.GetStatistics()
		require.NoError(t, err)
		require.NotNil(t, stats)
		require.NotNil(t, stats.Overall)

		assert.Equal(t, int64(0), stats.Overall.TotalPRs)
		assert.Equal(t, int64(0), stats.Overall.TotalAssignments)
		assert.Equal(t, int64(0), stats.Overall.TotalUsers)
		assert.Equal(t, int64(0), stats.Overall.TotalTeams)
		assert.Empty(t, stats.ReviewerStats)
		assert.Empty(t, stats.AuthorStats)
	})

	t.Run("success - statistics with data", func(t *testing.T) {
		// Setup: create teams, users, and PRs
		teamName1 := "team1"
		teamName2 := "team2"
		authorID := "author1"
		reviewerID1 := "reviewer1"
		reviewerID2 := "reviewer2"

		require.NoError(t, team.Create(db, teamName1))
		require.NoError(t, team.Create(db, teamName2))

		require.NoError(t, user.Create(db, &domain.User{
			UserID:   authorID,
			Username: "author",
			TeamName: teamName1,
			IsActive: true,
		}))
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   reviewerID1,
			Username: "reviewer1",
			TeamName: teamName1,
			IsActive: true,
		}))
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   reviewerID2,
			Username: "reviewer2",
			TeamName: teamName2,
			IsActive: true,
		}))

		// Create PRs
		prID1 := "pr1"
		prID2 := "pr2"
		prID3 := "pr3"

		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID1,
			PullRequestName: "PR 1",
			AuthorID:        authorID,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID2,
			PullRequestName: "PR 2",
			AuthorID:        authorID,
			Status:          domain.StatusOpen,
		}))
		require.NoError(t, pr.Create(db, &domain.PullRequest{
			PullRequestID:   prID3,
			PullRequestName: "PR 3",
			AuthorID:        reviewerID1,
			Status:          domain.StatusMerged,
		}))

		// Assign reviewers
		require.NoError(t, pr.InsertReviewer(db, prID1, reviewerID1))
		require.NoError(t, pr.InsertReviewer(db, prID1, reviewerID2))
		require.NoError(t, pr.InsertReviewer(db, prID2, reviewerID1))
		require.NoError(t, pr.InsertReviewer(db, prID3, reviewerID2))

		// Get statistics
		st, err := statsService.GetStatistics()
		require.NoError(t, err)
		require.NotNil(t, st)
		require.NotNil(t, st.Overall)

		// Check overall stats
		assert.Equal(t, int64(3), st.Overall.TotalPRs)
		assert.Equal(t, int64(4), st.Overall.TotalAssignments)
		assert.Equal(t, int64(3), st.Overall.TotalUsers)
		assert.Equal(t, int64(2), st.Overall.TotalTeams)

		// Check reviewer stats
		// reviewer1 should have 2 assignments, reviewer2 should have 2 assignments
		// author1 should have 0 assignments (only created PRs, not assigned as reviewer)
		assert.Len(t, st.ReviewerStats, 3)

		// Find reviewer1 in stats
		var reviewer1Stat, reviewer2Stat, authorStat *stats.ReviewerStat
		for i := range st.ReviewerStats {
			if st.ReviewerStats[i].UserID == reviewerID1 {
				reviewer1Stat = &st.ReviewerStats[i]
			}
			if st.ReviewerStats[i].UserID == reviewerID2 {
				reviewer2Stat = &st.ReviewerStats[i]
			}
			if st.ReviewerStats[i].UserID == authorID {
				authorStat = &st.ReviewerStats[i]
			}
		}

		require.NotNil(t, reviewer1Stat, "reviewer1 should be in stats")
		assert.Equal(t, "reviewer1", reviewer1Stat.Username)
		assert.Equal(t, int64(2), reviewer1Stat.Count)

		require.NotNil(t, reviewer2Stat, "reviewer2 should be in stats")
		assert.Equal(t, "reviewer2", reviewer2Stat.Username)
		assert.Equal(t, int64(2), reviewer2Stat.Count)

		require.NotNil(t, authorStat, "author should be in stats")
		assert.Equal(t, "author", authorStat.Username)
		assert.Equal(t, int64(0), authorStat.Count)

		// Check author stats
		// author1 should have 2 PRs, reviewer1 should have 1 PR
		assert.Len(t, st.AuthorStats, 3)

		// Find authors in stats
		var author1AuthorStat, reviewer1AuthorStat *stats.AuthorStat
		for i := range st.AuthorStats {
			if st.AuthorStats[i].UserID == authorID {
				author1AuthorStat = &st.AuthorStats[i]
			}
			if st.AuthorStats[i].UserID == reviewerID1 {
				reviewer1AuthorStat = &st.AuthorStats[i]
			}
		}

		require.NotNil(t, author1AuthorStat, "author1 should be in author stats")
		assert.Equal(t, "author", author1AuthorStat.Username)
		assert.Equal(t, int64(2), author1AuthorStat.Count)

		require.NotNil(t, reviewer1AuthorStat, "reviewer1 should be in author stats")
		assert.Equal(t, "reviewer1", reviewer1AuthorStat.Username)
		assert.Equal(t, int64(1), reviewer1AuthorStat.Count)
	})

	t.Run("success - user with no PRs or assignments", func(t *testing.T) {
		// Create a user with no PRs or assignments
		teamName := "team3"
		userID := "user_no_activity"

		require.NoError(t, team.Create(db, teamName))
		require.NoError(t, user.Create(db, &domain.User{
			UserID:   userID,
			Username: "no_activity",
			TeamName: teamName,
			IsActive: true,
		}))

		st, err := statsService.GetStatistics()
		require.NoError(t, err)

		// Find user in reviewer stats
		var userStat *stats.ReviewerStat
		for i := range st.ReviewerStats {
			if st.ReviewerStats[i].UserID == userID {
				userStat = &st.ReviewerStats[i]
				break
			}
		}

		require.NotNil(t, userStat)
		assert.Equal(t, int64(0), userStat.Count)

		// Find user in author stats
		var userAuthorStat *stats.AuthorStat
		for i := range st.AuthorStats {
			if st.AuthorStats[i].UserID == userID {
				userAuthorStat = &st.AuthorStats[i]
				break
			}
		}

		require.NotNil(t, userAuthorStat)
		assert.Equal(t, int64(0), userAuthorStat.Count)
	})
}

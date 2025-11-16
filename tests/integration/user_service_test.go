package integration

import (
	"database/sql"
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

func TestUserService_SetIsActive(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	// Setup: create team and user
	teamName := "test_team"
	userID := "user1"
	username := "test_user"

	require.NoError(t, team.Create(db, teamName))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   userID,
		Username: username,
		TeamName: teamName,
		IsActive: true,
	}))

	userService := service.NewUserService(db)

	tests := []struct {
		name           string
		userID         string
		isActive       bool
		expectedError  error
		expectedActive bool
	}{
		{
			name:           "success - activate user",
			userID:         userID,
			isActive:       true,
			expectedError:  nil,
			expectedActive: true,
		},
		{
			name:           "success - deactivate user",
			userID:         userID,
			isActive:       false,
			expectedError:  nil,
			expectedActive: false,
		},
		{
			name:          "error - user not found",
			userID:        "nonexistent",
			isActive:      true,
			expectedError: service.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := userService.SetIsActive(tt.userID, tt.isActive)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, assert.ErrorIs(t, err, tt.expectedError))
				assert.Nil(t, u)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, u)
				assert.Equal(t, tt.userID, u.UserID)
				assert.Equal(t, tt.expectedActive, u.IsActive)
			}
		})
	}
}

func TestUserService_GetUserReviews(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	defer func() { _ = tests.CleanupTestDB(db) }()

	// Setup: create team, users, and PR
	teamName := "test_team"
	authorID := "author1"
	reviewerID := "reviewer1"

	require.NoError(t, team.Create(db, teamName))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   authorID,
		Username: "author",
		TeamName: teamName,
		IsActive: true,
	}))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   reviewerID,
		Username: "reviewer",
		TeamName: teamName,
		IsActive: true,
	}))

	userService := service.NewUserService(db)

	t.Run("success - returns user reviews", func(t *testing.T) {
		// Create PR with reviewer
		prID := "pr1"
		prName := "Test PR"

		require.NoError(t, createPRWithReviewer(db, prID, prName, authorID, reviewerID))

		reviews, err := userService.GetUserReviews(reviewerID)
		require.NoError(t, err)
		assert.Len(t, reviews, 1)
		assert.Equal(t, prID, reviews[0].PullRequestID)
		assert.Equal(t, prName, reviews[0].PullRequestName)
		assert.Equal(t, authorID, reviews[0].AuthorID)
	})

	t.Run("success - empty reviews list", func(t *testing.T) {
		reviews, err := userService.GetUserReviews("user_with_no_reviews")
		require.NoError(t, err)
		assert.Empty(t, reviews)
	})

	t.Run("success - multiple reviews", func(t *testing.T) {
		// Create multiple PRs
		prID1 := "pr2"
		prID2 := "pr3"
		prName1 := "PR 2"
		prName2 := "PR 3"

		require.NoError(t, createPRWithReviewer(db, prID1, prName1, authorID, reviewerID))
		require.NoError(t, createPRWithReviewer(db, prID2, prName2, authorID, reviewerID))

		reviews, err := userService.GetUserReviews(reviewerID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(reviews), 2)
	})
}

// Helper function to create PR with reviewer
func createPRWithReviewer(db *sql.DB, prID, prName, authorID, reviewerID string) error {
	pullRequest := &domain.PullRequest{
		PullRequestID:   prID,
		PullRequestName: prName,
		AuthorID:        authorID,
		Status:          domain.StatusOpen,
	}
	if err := pr.Create(db, pullRequest); err != nil {
		return err
	}
	return pr.InsertReviewer(db, prID, reviewerID)
}

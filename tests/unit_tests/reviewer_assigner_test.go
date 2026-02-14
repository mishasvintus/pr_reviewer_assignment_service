package unit_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
)

func users(ids ...string) []domain.User {
	out := make([]domain.User, len(ids))
	for i, id := range ids {
		out[i] = domain.User{UserID: id, Username: id, TeamName: "t1", IsActive: true}
	}
	return out
}

func TestReviewerAssigner_SelectReviewers(t *testing.T) {
	assigner := service.NewReviewerAssigner()

	t.Run("empty teammates returns empty", func(t *testing.T) {
		got, err := assigner.SelectReviewers(nil)
		require.NoError(t, err)
		assert.Empty(t, got)

		got, err = assigner.SelectReviewers(users())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("one teammate returns one", func(t *testing.T) {
		teammates := users("u1")
		got, err := assigner.SelectReviewers(teammates)
		require.NoError(t, err)
		assert.Equal(t, []string{"u1"}, got)
	})

	t.Run("two teammates returns both", func(t *testing.T) {
		teammates := users("u1", "u2")
		got, err := assigner.SelectReviewers(teammates)
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.ElementsMatch(t, []string{"u1", "u2"}, got)
	})

	t.Run("three or more teammates returns two distinct from set", func(t *testing.T) {
		teammates := users("u1", "u2", "u3")
		got, err := assigner.SelectReviewers(teammates)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.NotEqual(t, got[0], got[1])
		ids := map[string]bool{"u1": true, "u2": true, "u3": true}
		assert.True(t, ids[got[0]], "first ID must be from teammates")
		assert.True(t, ids[got[1]], "second ID must be from teammates")
	})
}

func TestReviewerAssigner_SelectReassignReviewers(t *testing.T) {
	assigner := service.NewReviewerAssigner()

	t.Run("no candidates returns error", func(t *testing.T) {
		teammates := users("u1") // only u1, exclude u1 as author
		got, err := assigner.SelectReassignReviewers(teammates, "u1", nil)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "no candidates")
	})

	t.Run("excludes author", func(t *testing.T) {
		teammates := users("author", "r1", "r2")
		got, err := assigner.SelectReassignReviewers(teammates, "author", nil)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.NotContains(t, got, "author")
		assert.ElementsMatch(t, []string{"r1", "r2"}, got)
	})

	t.Run("excludes assigned reviewers", func(t *testing.T) {
		teammates := users("u1", "u2", "u3")
		got, err := assigner.SelectReassignReviewers(teammates, "author", []string{"u1", "u2"})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "u3", got[0])
	})

	t.Run("excludes author and assigned", func(t *testing.T) {
		teammates := users("a", "r1", "r2", "r3")
		got, err := assigner.SelectReassignReviewers(teammates, "a", []string{"r1"})
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.NotContains(t, got, "a")
		assert.NotContains(t, got, "r1")
		assert.ElementsMatch(t, got, []string{"r2", "r3"})
	})

	t.Run("one candidate returns one", func(t *testing.T) {
		teammates := users("author", "r1")
		got, err := assigner.SelectReassignReviewers(teammates, "author", nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"r1"}, got)
	})

	t.Run("all excluded returns error", func(t *testing.T) {
		teammates := users("u1", "u2")
		got, err := assigner.SelectReassignReviewers(teammates, "u1", []string{"u2"})
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

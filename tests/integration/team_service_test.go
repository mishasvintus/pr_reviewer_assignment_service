package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/team"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
	"github.com/mishasvintus/avito_backend_internship/tests"
)

func TestTeamService_CreateTeam(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer db.Close()
	defer tests.CleanupTestDB(db)

	teamService := service.NewTeamService(db)

	tests := []struct {
		name          string
		teamName      string
		members       []domain.TeamMember
		expectedError error
	}{
		{
			name:     "success - creates team with members",
			teamName: "team1",
			members: []domain.TeamMember{
				{UserID: "user1", Username: "user1", IsActive: true},
				{UserID: "user2", Username: "user2", IsActive: false},
			},
			expectedError: nil,
		},
		{
			name:          "success - creates team with no members",
			teamName:      "team2",
			members:       []domain.TeamMember{},
			expectedError: nil,
		},
		{
			name:     "error - team already exists",
			teamName: "team1",
			members: []domain.TeamMember{
				{UserID: "user3", Username: "user3", IsActive: true},
			},
			expectedError: service.ErrTeamExists,
		},
		{
			name:     "success - updates existing user to new team",
			teamName: "team3",
			members: []domain.TeamMember{
				{UserID: "user1", Username: "user1_updated", IsActive: true},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := teamService.CreateTeam(tt.teamName, tt.members)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, assert.ErrorIs(t, err, tt.expectedError))
			} else {
				assert.NoError(t, err)

				// Verify team was created
				team, err := team.Get(db, tt.teamName)
				require.NoError(t, err)
				assert.Equal(t, tt.teamName, team.TeamName)
				assert.Len(t, team.Members, len(tt.members))

				// Verify members (order may vary)
				memberMap := make(map[string]domain.TeamMember)
				for _, m := range team.Members {
					memberMap[m.UserID] = m
				}
				for _, expectedMember := range tt.members {
					actualMember, exists := memberMap[expectedMember.UserID]
					require.True(t, exists, "member %s not found", expectedMember.UserID)
					assert.Equal(t, expectedMember.Username, actualMember.Username)
					assert.Equal(t, expectedMember.IsActive, actualMember.IsActive)
				}
			}
		})
	}
}

func TestTeamService_GetTeam(t *testing.T) {
	db, err := tests.SetupTestDB()
	require.NoError(t, err)
	defer db.Close()
	defer tests.CleanupTestDB(db)

	// Setup: create team with members
	teamName := "test_team"
	require.NoError(t, team.Create(db, teamName))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   "user1",
		Username: "user1",
		TeamName: teamName,
		IsActive: true,
	}))
	require.NoError(t, user.Create(db, &domain.User{
		UserID:   "user2",
		Username: "user2",
		TeamName: teamName,
		IsActive: false,
	}))

	teamService := service.NewTeamService(db)

	tests := []struct {
		name          string
		teamName      string
		expectedError error
		validateTeam  func(*testing.T, *domain.Team)
	}{
		{
			name:          "success - returns team with members",
			teamName:      teamName,
			expectedError: nil,
			validateTeam: func(t *testing.T, team *domain.Team) {
				assert.Equal(t, teamName, team.TeamName)
				assert.Len(t, team.Members, 2)
			},
		},
		{
			name:          "success - returns team with no members",
			teamName:      "empty_team",
			expectedError: nil,
			validateTeam: func(t *testing.T, tm *domain.Team) {
				assert.Equal(t, "empty_team", tm.TeamName)
				assert.Empty(t, tm.Members)
			},
		},
		{
			name:          "error - team not found",
			teamName:      "nonexistent",
			expectedError: service.ErrTeamNotFound,
			validateTeam:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: create empty team if needed
			if tt.teamName == "empty_team" {
				require.NoError(t, team.Create(db, "empty_team"))
			}

			team, err := teamService.GetTeam(tt.teamName)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, assert.ErrorIs(t, err, tt.expectedError))
				assert.Nil(t, team)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, team)
				if tt.validateTeam != nil {
					tt.validateTeam(t, team)
				}
			}
		})
	}
}

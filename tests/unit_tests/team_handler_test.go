package unit_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/handler"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
	handlermocks "github.com/mishasvintus/avito_backend_internship/tests/mocks"
)

func TestTeamHandler_GetTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		queryParams      map[string]string
		mockSetup        func(*handlermocks.MockTeamServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns team with members",
			queryParams: map[string]string{
				"team_name": "team1",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				isActive1 := true
				isActive2 := false
				m.EXPECT().GetTeam("team1").Return(&domain.Team{
					TeamName: "team1",
					Members: []domain.TeamMember{
						{
							UserID:   "user1",
							Username: "Alice",
							IsActive: isActive1,
						},
						{
							UserID:   "user2",
							Username: "Bob",
							IsActive: isActive2,
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.TeamResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "team1", response.TeamName)
				assert.Len(t, response.Members, 2)
				assert.Equal(t, "user1", response.Members[0].UserID)
				assert.Equal(t, "Alice", response.Members[0].Username)
				assert.True(t, response.Members[0].IsActive)
				assert.Equal(t, "user2", response.Members[1].UserID)
				assert.False(t, response.Members[1].IsActive)
			},
		},
		{
			name: "success - returns team with no members",
			queryParams: map[string]string{
				"team_name": "empty_team",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().GetTeam("empty_team").Return(&domain.Team{
					TeamName: "empty_team",
					Members:  []domain.TeamMember{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.TeamResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "empty_team", response.TeamName)
				assert.Empty(t, response.Members)
			},
		},
		{
			name:           "error - missing team_name parameter",
			queryParams:    map[string]string{},
			mockSetup:      func(m *handlermocks.MockTeamServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "team_name parameter is required", response.Error.Message)
			},
		},
		{
			name: "error - empty team_name parameter",
			queryParams: map[string]string{
				"team_name": "",
			},
			mockSetup:      func(m *handlermocks.MockTeamServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "team_name parameter is required", response.Error.Message)
			},
		},
		{
			name: "error - team not found",
			queryParams: map[string]string{
				"team_name": "nonexistent",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().GetTeam("nonexistent").Return(nil, service.ErrTeamNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "team not found", response.Error.Message)
			},
		},
		{
			name: "error - internal error from service",
			queryParams: map[string]string{
				"team_name": "team1",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().GetTeam("team1").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error.Message, "assert.AnError")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := handlermocks.NewMockTeamServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewTeamHandler(mockService)

			req, err := http.NewRequest(http.MethodGet, "/team/get", nil)
			require.NoError(t, err)

			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.GetTeam(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

func TestTeamHandler_AddTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*handlermocks.MockTeamServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - creates team with members",
			requestBody: map[string]interface{}{
				"team_name": "team1",
				"members": []map[string]interface{}{
					{
						"user_id":   "user1",
						"username":  "Alice",
						"is_active": true,
					},
					{
						"user_id":   "user2",
						"username":  "Bob",
						"is_active": false,
					},
				},
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().CreateTeam("team1", []domain.TeamMember{
					{UserID: "user1", Username: "Alice", IsActive: true},
					{UserID: "user2", Username: "Bob", IsActive: false},
				}).Return(nil)

				m.EXPECT().GetTeam("team1").Return(&domain.Team{
					TeamName: "team1",
					Members: []domain.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
						{UserID: "user2", Username: "Bob", IsActive: false},
					},
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Team)
				assert.Equal(t, "team1", response.Team.TeamName)
				assert.Len(t, response.Team.Members, 2)
				assert.Equal(t, "user1", response.Team.Members[0].UserID)
				assert.Equal(t, "Alice", response.Team.Members[0].Username)
				assert.True(t, response.Team.Members[0].IsActive)
			},
		},
		{
			name: "success - creates team with empty members",
			requestBody: map[string]interface{}{
				"team_name": "empty_team",
				"members":   []map[string]interface{}{},
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().CreateTeam("empty_team", []domain.TeamMember{}).Return(nil)
				m.EXPECT().GetTeam("empty_team").Return(&domain.Team{
					TeamName: "empty_team",
					Members:  []domain.TeamMember{},
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Team)
				assert.Equal(t, "empty_team", response.Team.TeamName)
				assert.Empty(t, response.Team.Members)
			},
		},
		{
			name: "error - invalid request body",
			requestBody: map[string]interface{}{
				"team_name": "team1",
				// missing members
			},
			mockSetup:      func(m *handlermocks.MockTeamServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "invalid request body", response.Error.Message)
			},
		},
		{
			name: "error - team already exists",
			requestBody: map[string]interface{}{
				"team_name": "existing_team",
				"members": []map[string]interface{}{
					{"user_id": "user1", "username": "Alice", "is_active": true},
				},
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().CreateTeam("existing_team", []domain.TeamMember{
					{UserID: "user1", Username: "Alice", IsActive: true},
				}).Return(service.ErrTeamExists)
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorTeamExists, response.Error.Code)
				assert.Equal(t, "team_name already exists", response.Error.Message)
			},
		},
		{
			name: "error - internal error from CreateTeam",
			requestBody: map[string]interface{}{
				"team_name": "team1",
				"members": []map[string]interface{}{
					{"user_id": "user1", "username": "Alice", "is_active": true},
				},
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().CreateTeam("team1", []domain.TeamMember{
					{UserID: "user1", Username: "Alice", IsActive: true},
				}).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error.Message, "assert.AnError")
			},
		},
		{
			name: "error - failed to retrieve created team",
			requestBody: map[string]interface{}{
				"team_name": "team1",
				"members": []map[string]interface{}{
					{"user_id": "user1", "username": "Alice", "is_active": true},
				},
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().CreateTeam("team1", []domain.TeamMember{
					{UserID: "user1", Username: "Alice", IsActive: true},
				}).Return(nil)
				m.EXPECT().GetTeam("team1").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "failed to retrieve created team", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := handlermocks.NewMockTeamServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewTeamHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.AddTeam(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

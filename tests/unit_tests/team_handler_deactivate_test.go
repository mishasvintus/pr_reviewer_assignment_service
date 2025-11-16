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

	"github.com/mishasvintus/avito_backend_internship/internal/handler"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
	handlermocks "github.com/mishasvintus/avito_backend_internship/tests/mocks"
)

func TestTeamHandler_DeactivateTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*handlermocks.MockTeamServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - team deactivated",
			requestBody: map[string]interface{}{
				"team_name": "test_team",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().DeactivateTeam("test_team").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "team deactivated successfully", response["message"])
			},
		},
		{
			name:        "error - invalid request body (missing team_name)",
			requestBody: map[string]interface{}{
				// missing team_name
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
			name: "error - team not found",
			requestBody: map[string]interface{}{
				"team_name": "nonexistent_team",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().DeactivateTeam("nonexistent_team").Return(service.ErrTeamNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "NOT_FOUND", string(response.Error.Code))
				assert.Equal(t, "team not found", response.Error.Message)
			},
		},
		{
			name: "error - internal server error",
			requestBody: map[string]interface{}{
				"team_name": "error_team",
			},
			mockSetup: func(m *handlermocks.MockTeamServiceInterface) {
				m.EXPECT().DeactivateTeam("error_team").Return(assert.AnError)
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

			teamHandler := handler.NewTeamHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			teamHandler.DeactivateTeam(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

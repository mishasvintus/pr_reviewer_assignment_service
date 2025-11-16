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

func TestUserHandler_SetIsActive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*handlermocks.MockUserServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - activate user",
			requestBody: map[string]interface{}{
				"user_id":   "user1",
				"is_active": true,
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().SetIsActive("user1", true).Return(&domain.User{
					UserID:   "user1",
					Username: "testuser",
					TeamName: "team1",
					IsActive: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.User)
				assert.Equal(t, "user1", response.User.UserID)
				assert.Equal(t, "testuser", response.User.Username)
				assert.Equal(t, "team1", response.User.TeamName)
				assert.True(t, response.User.IsActive)
			},
		},
		{
			name: "success - deactivate user",
			requestBody: map[string]interface{}{
				"user_id":   "user1",
				"is_active": false,
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().SetIsActive("user1", false).Return(&domain.User{
					UserID:   "user1",
					Username: "testuser",
					TeamName: "team1",
					IsActive: false,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.User)
				assert.False(t, response.User.IsActive)
			},
		},
		{
			name: "error - invalid request body",
			requestBody: map[string]interface{}{
				"user_id": "user1",
				// missing is_active
			},
			mockSetup:      func(m *handlermocks.MockUserServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "invalid request body", response.Error.Message)
			},
		},
		{
			name: "error - user not found",
			requestBody: map[string]interface{}{
				"user_id":   "nonexistent",
				"is_active": true,
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().SetIsActive("nonexistent", true).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user not found", response.Error.Message)
			},
		},
		{
			name: "error - internal error",
			requestBody: map[string]interface{}{
				"user_id":   "user1",
				"is_active": true,
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().SetIsActive("user1", true).Return(nil, assert.AnError)
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
			mockService := handlermocks.NewMockUserServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewUserHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.SetIsActive(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

func TestUserHandler_GetReview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		queryParams      map[string]string
		mockSetup        func(*handlermocks.MockUserServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns user reviews",
			queryParams: map[string]string{
				"user_id": "user1",
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().GetUserReviews("user1").Return([]domain.PullRequestShort{
					{
						PullRequestID:   "pr1",
						PullRequestName: "Fix bug",
						AuthorID:        "author1",
						Status:          domain.StatusOpen,
					},
					{
						PullRequestID:   "pr2",
						PullRequestName: "Add feature",
						AuthorID:        "author2",
						Status:          domain.StatusMerged,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.GetReviewResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user1", response.UserID)
				assert.Len(t, response.PullRequests, 2)
				assert.Equal(t, "pr1", response.PullRequests[0].PullRequestID)
				assert.Equal(t, "Fix bug", response.PullRequests[0].PullRequestName)
				assert.Equal(t, "author1", response.PullRequests[0].AuthorID)
				assert.Equal(t, "OPEN", response.PullRequests[0].Status)
				assert.Equal(t, "pr2", response.PullRequests[1].PullRequestID)
				assert.Equal(t, "MERGED", response.PullRequests[1].Status)
			},
		},
		{
			name: "success - empty reviews list",
			queryParams: map[string]string{
				"user_id": "user1",
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().GetUserReviews("user1").Return([]domain.PullRequestShort{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.GetReviewResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user1", response.UserID)
				assert.Empty(t, response.PullRequests)
			},
		},
		{
			name:           "error - missing user_id parameter",
			queryParams:    map[string]string{},
			mockSetup:      func(m *handlermocks.MockUserServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user_id parameter is required", response.Error.Message)
			},
		},
		{
			name: "error - empty user_id parameter",
			queryParams: map[string]string{
				"user_id": "",
			},
			mockSetup:      func(m *handlermocks.MockUserServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user_id parameter is required", response.Error.Message)
			},
		},
		{
			name: "error - internal error from service",
			queryParams: map[string]string{
				"user_id": "user1",
			},
			mockSetup: func(m *handlermocks.MockUserServiceInterface) {
				m.EXPECT().GetUserReviews("user1").Return(nil, assert.AnError)
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
			mockService := handlermocks.NewMockUserServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewUserHandler(mockService)

			req, err := http.NewRequest(http.MethodGet, "/users/getReview", nil)
			require.NoError(t, err)

			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.GetReview(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

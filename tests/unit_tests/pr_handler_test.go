package unit_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/handler"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
	handlermocks "github.com/mishasvintus/avito_backend_internship/tests/mocks"
)

func TestPRHandler_MergePR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	mergedAt := now.Add(1 * time.Hour)

	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*handlermocks.MockPRServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - merges PR",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().MergePR("pr1").Return(&domain.PullRequest{
					PullRequestID:     "pr1",
					PullRequestName:   "Fix bug",
					AuthorID:          "author1",
					Status:            domain.StatusMerged,
					AssignedReviewers: []string{"reviewer1"},
					CreatedAt:         &now,
					MergedAt:          &mergedAt,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.PR)
				assert.Equal(t, "pr1", response.PR.PullRequestID)
				assert.Equal(t, "Fix bug", response.PR.PullRequestName)
				assert.Equal(t, "MERGED", response.PR.Status)
				assert.NotEmpty(t, response.PR.MergedAt)
			},
		},
		{
			name:        "error - invalid request body",
			requestBody: map[string]interface{}{
				// missing pull_request_id
			},
			mockSetup:      func(m *handlermocks.MockPRServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "invalid request body", response.Error.Message)
			},
		},
		{
			name: "error - PR not found",
			requestBody: map[string]interface{}{
				"pull_request_id": "nonexistent",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().MergePR("nonexistent").Return(nil, service.ErrPRNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "pull request not found", response.Error.Message)
			},
		},
		{
			name: "error - internal error from service",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().MergePR("pr1").Return(nil, assert.AnError)
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
			mockService := handlermocks.NewMockPRServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewPRHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.MergePR(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

func TestPRHandler_CreatePR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()

	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*handlermocks.MockPRServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - creates PR with reviewers",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr1",
				"pull_request_name": "Fix bug",
				"author_id":         "author1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().CreatePR("pr1", "Fix bug", "author1").Return(&domain.PullRequest{
					PullRequestID:     "pr1",
					PullRequestName:   "Fix bug",
					AuthorID:          "author1",
					Status:            domain.StatusOpen,
					AssignedReviewers: []string{"reviewer1", "reviewer2"},
					CreatedAt:         &now,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.PR)
				assert.Equal(t, "pr1", response.PR.PullRequestID)
				assert.Equal(t, "Fix bug", response.PR.PullRequestName)
				assert.Equal(t, "author1", response.PR.AuthorID)
				assert.Equal(t, "OPEN", response.PR.Status)
				assert.Len(t, response.PR.AssignedReviewers, 2)
			},
		},
		{
			name: "error - invalid request body",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				// missing pull_request_name and author_id
			},
			mockSetup:      func(m *handlermocks.MockPRServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "invalid request body", response.Error.Message)
			},
		},
		{
			name: "error - PR already exists",
			requestBody: map[string]interface{}{
				"pull_request_id":   "existing_pr",
				"pull_request_name": "Fix bug",
				"author_id":         "author1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().CreatePR("existing_pr", "Fix bug", "author1").Return(nil, service.ErrPRExists)
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorPRExists, response.Error.Code)
				assert.Equal(t, "PR id already exists", response.Error.Message)
			},
		},
		{
			name: "error - author or team not found",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr1",
				"pull_request_name": "Fix bug",
				"author_id":         "nonexistent",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().CreatePR("pr1", "Fix bug", "nonexistent").Return(nil, service.ErrPRAuthorNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "author or team not found", response.Error.Message)
			},
		},
		{
			name: "error - inactive reviewer",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr1",
				"pull_request_name": "Fix bug",
				"author_id":         "author1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().CreatePR("pr1", "Fix bug", "author1").Return(nil, service.ErrInactiveReviewer)
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, service.ErrInactiveReviewer.Error(), response.Error.Message)
			},
		},
		{
			name: "error - internal error from service",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr1",
				"pull_request_name": "Fix bug",
				"author_id":         "author1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().CreatePR("pr1", "Fix bug", "author1").Return(nil, assert.AnError)
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
			mockService := handlermocks.NewMockPRServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewPRHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreatePR(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

func TestPRHandler_ReassignPR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()

	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*handlermocks.MockPRServiceInterface)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - reassigns reviewer",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				"old_user_id":     "old_reviewer",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("pr1", "old_reviewer").Return(&domain.PullRequest{
					PullRequestID:     "pr1",
					PullRequestName:   "Fix bug",
					AuthorID:          "author1",
					Status:            domain.StatusOpen,
					AssignedReviewers: []string{"new_reviewer"},
					CreatedAt:         &now,
				}, "new_reviewer", nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ReassignResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.PR)
				assert.Equal(t, "pr1", response.PR.PullRequestID)
				assert.Equal(t, "new_reviewer", response.ReplacedBy)
				assert.Contains(t, response.PR.AssignedReviewers, "new_reviewer")
			},
		},
		{
			name: "error - invalid request body",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				// missing old_user_id
			},
			mockSetup:      func(m *handlermocks.MockPRServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "invalid request body", response.Error.Message)
			},
		},
		{
			name: "error - PR not found",
			requestBody: map[string]interface{}{
				"pull_request_id": "nonexistent",
				"old_user_id":     "reviewer1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("nonexistent", "reviewer1").Return(nil, "", service.ErrPRNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "pull request or user not found", response.Error.Message)
			},
		},
		{
			name: "error - author not found",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				"old_user_id":     "reviewer1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("pr1", "reviewer1").Return(nil, "", service.ErrPRAuthorNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "pull request or user not found", response.Error.Message)
			},
		},
		{
			name: "error - PR already merged",
			requestBody: map[string]interface{}{
				"pull_request_id": "merged_pr",
				"old_user_id":     "reviewer1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("merged_pr", "reviewer1").Return(nil, "", service.ErrPRMerged)
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorPRMerged, response.Error.Code)
				assert.Equal(t, "cannot reassign on merged PR", response.Error.Message)
			},
		},
		{
			name: "error - reviewer not assigned",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				"old_user_id":     "not_assigned",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("pr1", "not_assigned").Return(nil, "", service.ErrReviewerNotAssigned)
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorNotAssigned, response.Error.Code)
				assert.Equal(t, "reviewer is not assigned to this PR", response.Error.Message)
			},
		},
		{
			name: "error - no candidate available",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				"old_user_id":     "reviewer1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("pr1", "reviewer1").Return(nil, "", service.ErrNoCandidate)
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, handler.ErrorNoCandidate, response.Error.Code)
				assert.Equal(t, "no active replacement candidate in team", response.Error.Message)
			},
		},
		{
			name: "error - inactive reviewer",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				"old_user_id":     "reviewer1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("pr1", "reviewer1").Return(nil, "", service.ErrInactiveReviewer)
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, service.ErrInactiveReviewer.Error(), response.Error.Message)
			},
		},
		{
			name: "error - internal error from service",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr1",
				"old_user_id":     "reviewer1",
			},
			mockSetup: func(m *handlermocks.MockPRServiceInterface) {
				m.EXPECT().ReassignPR("pr1", "reviewer1").Return(nil, "", assert.AnError)
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
			mockService := handlermocks.NewMockPRServiceInterface(t)
			tt.mockSetup(mockService)

			handler := handler.NewPRHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ReassignPR(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

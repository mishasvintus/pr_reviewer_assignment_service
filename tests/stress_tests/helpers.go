package stress_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mishasvintus/avito_backend_internship/internal/handler"
	"github.com/mishasvintus/avito_backend_internship/tests"
)

// setupTestData creates test data for basic load tests.
func setupTestData(t *testing.T) {
	// Clean up database before setting up test data
	db, err := tests.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create team
	teamData := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "User1", "is_active": true},
			{"user_id": "u2", "username": "User2", "is_active": true},
			{"user_id": "u3", "username": "User3", "is_active": true},
		},
	}
	jsonData, _ := json.Marshal(teamData)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to setup team: %v", err)
	}
	_ = resp.Body.Close()

	// Create some PRs for testing
	for i := 0; i < 5; i++ {
		prID := fmt.Sprintf("pr-load-%d", i)
		prData := map[string]interface{}{
			"pull_request_id":   prID,
			"pull_request_name": fmt.Sprintf("Load Test PR %d", i),
			"author_id":         "u1",
		}
		jsonData, _ := json.Marshal(prData)
		resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create PR %s: %v", prID, err)
		}
		_ = resp.Body.Close()
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)
}

// setupReassignTestData creates test data for reassign load test.
func setupReassignTestData(t *testing.T) ([]string, []string) {
	// Clean up database
	db, err := tests.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create team with 10 members
	members := make([]map[string]interface{}, 10)
	userIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("u%d", i+1)
		userIDs[i] = userID
		members[i] = map[string]interface{}{
			"user_id":   userID,
			"username":  fmt.Sprintf("User%d", i+1),
			"is_active": true,
		}
	}

	teamData := map[string]interface{}{
		"team_name": "reassign-team",
		"members":   members,
	}
	jsonData, _ := json.Marshal(teamData)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to setup team: %v", err)
	}
	_ = resp.Body.Close()

	// Create 10 PRs with different authors
	prIDs := make([]string, 10)
	initialReviewerIDs := make([]string, 10)

	for i := 0; i < 10; i++ {
		prID := fmt.Sprintf("pr-reassign-%d", i)
		prIDs[i] = prID
		authorID := userIDs[i]

		prData := map[string]interface{}{
			"pull_request_id":   prID,
			"pull_request_name": fmt.Sprintf("Reassign Test PR %d", i),
			"author_id":         authorID,
		}
		jsonData, _ := json.Marshal(prData)
		resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create PR %s: %v", prID, err)
		}

		// Parse response to get initial reviewers
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		var createResp struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
			} `json:"pr"`
		}
		if json.Unmarshal(body, &createResp) == nil && len(createResp.PR.AssignedReviewers) > 0 {
			// Store initial reviewers
			initialReviewerIDs[i] = createResp.PR.AssignedReviewers[0]
			// Initialize currentReviewers map
			currentReviewers.Store(prID, createResp.PR.AssignedReviewers)
		} else {
			// Fallback - use next user as initial reviewer
			initialReviewerIDs[i] = userIDs[(i+1)%10]
			currentReviewers.Store(prID, []string{initialReviewerIDs[i]})
		}

		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)
	return prIDs, initialReviewerIDs
}

// testGetTeam performs a GET /team/get request.
func testGetTeam(results chan<- Result) {
	defer func() {
		_ = recover()
	}()

	start := time.Now()
	resp, err := http.Get(baseURL + "/team/get?team_name=" + teamName)
	duration := time.Since(start)

	result := Result{Endpoint: "GET /team/get", Duration: duration}
	if err != nil {
		result.Error = err
	} else {
		defer func() { _ = resp.Body.Close() }()
		_, _ = io.Copy(io.Discard, resp.Body)
		result.StatusCode = resp.StatusCode
	}

	select {
	case results <- result:
	default:
	}
}

// testGetUserReviews performs a GET /users/getReview request.
func testGetUserReviews(results chan<- Result) {
	defer func() {
		_ = recover()
	}()

	start := time.Now()
	resp, err := http.Get(baseURL + "/users/getReview?user_id=u2")
	duration := time.Since(start)

	result := Result{Endpoint: "GET /users/getReview", Duration: duration}
	if err != nil {
		result.Error = err
	} else {
		defer func() { _ = resp.Body.Close() }()
		_, _ = io.Copy(io.Discard, resp.Body)
		result.StatusCode = resp.StatusCode
	}

	select {
	case results <- result:
	default:
	}
}

// testSetIsActive performs a POST /users/setIsActive request.
func testSetIsActive(results chan<- Result) {
	defer func() {
		_ = recover()
	}()

	data := map[string]interface{}{
		"user_id":   "u3",
		"is_active": true,
	}
	jsonData, _ := json.Marshal(data)

	start := time.Now()
	resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(jsonData))
	duration := time.Since(start)

	result := Result{Endpoint: "POST /users/setIsActive", Duration: duration}
	if err != nil {
		result.Error = err
	} else {
		defer func() { _ = resp.Body.Close() }()
		_, _ = io.Copy(io.Discard, resp.Body)
		result.StatusCode = resp.StatusCode
	}

	select {
	case results <- result:
	default:
	}
}

// testReassignPR performs a POST /pullRequest/reassign request.
func testReassignPR(results chan<- Result, prIDs []string, reviewerIDs []string) {
	defer func() {
		_ = recover()
	}()

	// Use round-robin to cycle through PRs
	counter := atomic.AddInt64(&reassignCounter, 1)
	prIndex := int(counter-1) % len(prIDs)
	prID := prIDs[prIndex]

	// Get current reviewers for this PR
	reviewerMutex.Lock()
	var currentReviewersList []string
	if val, ok := currentReviewers.Load(prID); ok {
		currentReviewersList = val.([]string)
	} else {
		// First time - use initial reviewers
		currentReviewersList = []string{reviewerIDs[prIndex]}
		currentReviewers.Store(prID, currentReviewersList)
	}
	reviewerMutex.Unlock()

	if len(currentReviewersList) == 0 {
		return
	}

	// Pick first reviewer to reassign
	oldReviewerID := currentReviewersList[0]

	data := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     oldReviewerID,
	}
	jsonData, _ := json.Marshal(data)

	start := time.Now()
	resp, err := http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(jsonData))
	duration := time.Since(start)

	result := Result{Endpoint: "POST /pullRequest/reassign", Duration: duration}
	if err != nil {
		result.Error = err
	} else {
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			var reassignResp handler.ReassignResponse
			if json.Unmarshal(body, &reassignResp) == nil && reassignResp.PR != nil {
				// Update current reviewers for this PR
				reviewerMutex.Lock()
				currentReviewers.Store(prID, reassignResp.PR.AssignedReviewers)
				reviewerMutex.Unlock()
			}
		} else {
			_, _ = io.Copy(io.Discard, resp.Body)
		}
		result.StatusCode = resp.StatusCode
	}

	select {
	case results <- result:
	default:
	}
}

// analyzeResults analyzes test results and prints statistics.
func analyzeResults(t *testing.T, results []Result, totalTime time.Duration) {
	if len(results) == 0 {
		t.Error("No results collected")
		return
	}

	var successCount, errorCount int64
	var totalDuration time.Duration
	var maxDuration, minDuration time.Duration = 0, time.Hour

	statusCodes := make(map[int]int64)
	endpointStats := make(map[string]struct {
		count    int64
		duration time.Duration
		errors   int64
	})

	for _, r := range results {
		statusCodes[r.StatusCode]++

		stats := endpointStats[r.Endpoint]
		stats.count++

		if r.Error != nil || r.StatusCode != 200 {
			errorCount++
			stats.errors++
		} else {
			successCount++
		}

		if r.Duration > 0 {
			totalDuration += r.Duration
			stats.duration += r.Duration
			if r.Duration > maxDuration {
				maxDuration = r.Duration
			}
			if r.Duration < minDuration {
				minDuration = r.Duration
			}
		}

		endpointStats[r.Endpoint] = stats
	}

	avgDuration := totalDuration / time.Duration(len(results))
	successRate := float64(successCount) / float64(len(results)) * 100
	actualRPS := float64(len(results)) / totalTime.Seconds()

	t.Logf("Total requests: %d", len(results))
	t.Logf("Total duration: %.2fs", totalTime.Seconds())
	t.Logf("Actual RPS: %.2f", actualRPS)
	t.Logf("Success: %d (%.2f%%)", successCount, successRate)
	t.Logf("Errors: %d (%.2f%%)", errorCount, 100-successRate)

	t.Logf("\nStatus codes:")
	for code, count := range statusCodes {
		t.Logf("  %d: %d", code, count)
	}

	t.Logf("\nResponse times:")
	t.Logf("  Average: %v", avgDuration)
	t.Logf("  Min: %v", minDuration)
	t.Logf("  Max: %v", maxDuration)

	t.Logf("\nPer endpoint:")
	for endpoint, stats := range endpointStats {
		avg := stats.duration / time.Duration(stats.count)
		epSuccessRate := float64(stats.count-stats.errors) / float64(stats.count) * 100
		t.Logf("  %s:", endpoint)
		t.Logf("    Requests: %d", stats.count)
		t.Logf("    Success rate: %.2f%%", epSuccessRate)
		t.Logf("    Avg response: %v", avg)
	}

	t.Logf("\nSLI Check:")
	t.Logf("  Success rate >= 99.9%%: %v (%.2f%%)", successRate >= 99.9, successRate)
	t.Logf("  Avg response <= 300ms: %v (%v)", avgDuration <= 300*time.Millisecond, avgDuration)

	// Проверяем SLI требования
	if successRate < 99.9 {
		t.Errorf("SLI requirement not met: success rate %.2f%% < 99.9%%", successRate)
	}
	if avgDuration > 300*time.Millisecond {
		t.Errorf("SLI requirement not met: avg response time %v > 300ms", avgDuration)
	}

	if successRate >= 99.9 && avgDuration <= 300*time.Millisecond {
		t.Logf("\nAll SLI requirements met")
	} else {
		t.Logf("\nSome SLI requirements not met")
	}
}

// calculatePercentile calculates the percentile value from sorted durations.
func calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	index := int(float64(len(durations)) * percentile / 100.0)
	if index >= len(durations) {
		index = len(durations) - 1
	}
	return durations[index]
}

// analyzeResultsWithPercentiles analyzes test results with percentile statistics.
func analyzeResultsWithPercentiles(t *testing.T, results []Result, totalTime time.Duration) {
	if len(results) == 0 {
		t.Error("No results collected")
		return
	}

	var successCount, errorCount int64
	var totalDuration time.Duration
	var maxDuration, minDuration time.Duration = 0, time.Hour

	statusCodes := make(map[int]int64)
	durations := make([]time.Duration, 0, len(results))

	for _, r := range results {
		statusCodes[r.StatusCode]++

		if r.Error != nil || r.StatusCode != 200 {
			errorCount++
		} else {
			successCount++
		}

		if r.Duration > 0 {
			durations = append(durations, r.Duration)
			totalDuration += r.Duration
			if r.Duration > maxDuration {
				maxDuration = r.Duration
			}
			if r.Duration < minDuration {
				minDuration = r.Duration
			}
		}
	}

	slices.Sort(durations)

	avgDuration := totalDuration / time.Duration(len(durations))
	successRate := float64(successCount) / float64(len(results)) * 100
	actualRPS := float64(len(results)) / totalTime.Seconds()

	p50 := calculatePercentile(durations, 50)
	p95 := calculatePercentile(durations, 95)
	p99 := calculatePercentile(durations, 99)

	t.Logf("Total requests: %d", len(results))
	t.Logf("Total duration: %.2fs", totalTime.Seconds())
	t.Logf("Actual RPS: %.2f", actualRPS)
	t.Logf("Success: %d (%.2f%%)", successCount, successRate)
	t.Logf("Errors: %d (%.2f%%)", errorCount, 100-successRate)

	t.Logf("\nStatus codes:")
	for code, count := range statusCodes {
		t.Logf("  %d: %d", code, count)
	}

	t.Logf("\nResponse times:")
	t.Logf("  Average: %v", avgDuration)
	t.Logf("  Min: %v", minDuration)
	t.Logf("  Max: %v", maxDuration)
	t.Logf("  p50: %v", p50)
	t.Logf("  p95: %v", p95)
	t.Logf("  p99: %v", p99)

	t.Logf("\nSLI Check:")
	t.Logf("  Success rate >= 99.9%%: %v (%.2f%%)", successRate >= 99.9, successRate)
	t.Logf("  Avg response <= 300ms: %v (%v)", avgDuration <= 300*time.Millisecond, avgDuration)

	// SLI
	if successRate < 99.9 {
		t.Errorf("SLI requirement not met: success rate %.2f%% < 99.9%%", successRate)
	}
	if avgDuration > 300*time.Millisecond {
		t.Errorf("SLI requirement not met: avg response time %v > 300ms", avgDuration)
	}
	if p95 > 500*time.Millisecond {
		t.Logf("Warning: p95 response time %v > 500ms", p95)
	}
	if p99 > 1000*time.Millisecond {
		t.Logf("Warning: p99 response time %v > 1000ms", p99)
	}

	if successRate >= 99.9 && avgDuration <= 300*time.Millisecond {
		t.Logf("\nAll SLI requirements met!")
	} else {
		t.Logf("\nSome SLI requirements not met")
	}
}

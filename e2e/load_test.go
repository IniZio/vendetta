package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const coordServerURL = "http://localhost:3001"

type workspaceCreateReq struct {
	GitHubUsername string `json:"github_username"`
	WorkspaceName  string `json:"workspace_name"`
	Repo           struct {
		Owner  string `json:"owner"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Branch string `json:"branch"`
		IsFork bool   `json:"is_fork"`
	} `json:"repo"`
	Provider string        `json:"provider"`
	Image    string        `json:"image"`
	Services []interface{} `json:"services"`
}

type workspaceCreateResp struct {
	WorkspaceID string `json:"workspace_id"`
	Status      string `json:"status"`
}

func createWorkspaceHTTP(idx int) (string, error) {
	req := workspaceCreateReq{
		GitHubUsername: fmt.Sprintf("loadtest-user%d", idx),
		WorkspaceName:  fmt.Sprintf("workspace-%d", idx),
		Provider:       "lxc",
		Image:          "ubuntu:22.04",
		Services:       []interface{}{},
	}
	req.Repo.Owner = "oursky"
	req.Repo.Name = "epson-eshop"
	req.Repo.URL = "https://github.com/oursky/epson-eshop.git"
	req.Repo.Branch = "main"
	req.Repo.IsFork = false

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/workspaces/create-from-repo", coordServerURL), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, string(respBody))
	}

	var resp workspaceCreateResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", err
	}

	return resp.WorkspaceID, nil
}

func TestConcurrentWorkspaceCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	numWorkspaces := 10
	var successCount, failureCount int32
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < numWorkspaces; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			wsID, err := createWorkspaceHTTP(idx)
			if err != nil {
				atomic.AddInt32(&failureCount, 1)
				return
			}

			if wsID != "" {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failureCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Concurrent workspace creation: %d success, %d failures in %v",
		successCount, failureCount, duration)

	assert.Less(t, duration, 30*time.Second, "should create workspaces within 30 seconds")
}

func TestWorkspaceCreationThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	const duration = 10 * time.Second
	var totalCount, failureCount int32

	startTime := time.Now()
	deadline := startTime.Add(duration)

	var wg sync.WaitGroup
	for worker := 0; worker < 5; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			idx := 0
			for time.Now().Before(deadline) {
				wsID, err := createWorkspaceHTTP(id*1000 + idx)
				if err == nil && wsID != "" {
					atomic.AddInt32(&totalCount, 1)
				} else {
					atomic.AddInt32(&failureCount, 1)
				}
				idx++
			}
		}(worker)
	}

	wg.Wait()
	actualDuration := time.Since(startTime)

	throughput := float64(totalCount) / actualDuration.Seconds()
	failureRate := float64(failureCount) / float64(totalCount+failureCount)

	t.Logf("Throughput: %.2f workspaces/sec, failure rate: %.1f%%", throughput, failureRate*100)

	assert.Greater(t, throughput, 0.5, "should create at least 0.5 workspaces per second")
}

func TestWorkspaceListingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	timings := make([]time.Duration, 0, 100)
	for i := 0; i < 100; i++ {
		start := time.Now()

		httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/workspaces", coordServerURL), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		httpResp, err := client.Do(httpReq)
		if err != nil {
			continue
		}
		io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		timings = append(timings, time.Since(start))
	}

	totalTime := time.Duration(0)
	for _, d := range timings {
		totalTime += d
	}
	avgTime := totalTime / time.Duration(len(timings))

	t.Logf("List workspaces: avg %.2f ms (100 requests)", float64(avgTime.Milliseconds()))
	assert.Less(t, avgTime, 500*time.Millisecond, "list should complete within 500ms")
}

func BenchmarkWorkspaceCreation(b *testing.B) {
	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := workspaceCreateReq{
			GitHubUsername: fmt.Sprintf("bench-user-%d", i),
			WorkspaceName:  fmt.Sprintf("bench-ws-%d", i),
			Provider:       "lxc",
			Image:          "ubuntu:22.04",
			Services:       []interface{}{},
		}
		req.Repo.Owner = "oursky"
		req.Repo.Name = "epson-eshop"
		req.Repo.URL = "https://github.com/oursky/epson-eshop.git"
		req.Repo.Branch = "main"
		req.Repo.IsFork = false

		body, _ := json.Marshal(req)

		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/workspaces/create-from-repo", coordServerURL), bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		httpResp, err := client.Do(httpReq)
		if err == nil {
			io.ReadAll(httpResp.Body)
			httpResp.Body.Close()
		}
	}
}

package testfixtures

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestEnvironment manages the complete test setup
type TestEnvironment struct {
	t           *testing.T
	baseDir     string
	gitRepoDir  string
	vendattaBin string
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	baseDir := t.TempDir()

	// Build vendatta binary
	vendattaBin := filepath.Join(baseDir, "vendatta")
	buildVendatta(t, vendattaBin)

	return &TestEnvironment{
		t:           t,
		baseDir:     baseDir,
		vendattaBin: vendattaBin,
	}
}

// SetupTestRepo creates a test git repository with realistic project structure
func (te *TestEnvironment) SetupTestRepo() string {
	te.gitRepoDir = filepath.Join(te.baseDir, "test-repo")
	te.runCommand("git", "init", te.gitRepoDir)
	te.runCommandInDir(te.gitRepoDir, "git", "config", "user.name", "Test User")
	te.runCommandInDir(te.gitRepoDir, "git", "config", "user.email", "test@example.com")

	// Create realistic project structure
	te.createFile(filepath.Join(te.gitRepoDir, "package.json"), `{
  "name": "test-project",
  "version": "1.0.0",
  "scripts": {
    "dev": "node server.js"
  }
}`)

	te.createFile(filepath.Join(te.gitRepoDir, "server.js"), `
const http = require('http');
const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from test server\n');
});
server.listen(3000, '0.0.0.0', () => {
  console.log('Server running on port 3000');
});
`)

	te.createFile(filepath.Join(te.gitRepoDir, "README.md"), "# Test Project\n\nA test project for e2e testing.")

	// Create initial commit
	te.runCommandInDir(te.gitRepoDir, "git", "add", ".")
	te.runCommandInDir(te.gitRepoDir, "git", "commit", "-m", "Initial commit")

	// Create test branches
	te.runCommandInDir(te.gitRepoDir, "git", "checkout", "-b", "feature-1")
	te.createFile(filepath.Join(te.gitRepoDir, "feature1.txt"), "Feature 1 content")
	te.runCommandInDir(te.gitRepoDir, "git", "add", ".")
	te.runCommandInDir(te.gitRepoDir, "git", "commit", "-m", "Add feature 1")

	te.runCommandInDir(te.gitRepoDir, "git", "checkout", "-b", "feature-2")
	te.createFile(filepath.Join(te.gitRepoDir, "feature2.txt"), "Feature 2 content")
	te.runCommandInDir(te.gitRepoDir, "git", "add", ".")
	te.runCommandInDir(te.gitRepoDir, "git", "commit", "-m", "Add feature 2")

	te.runCommandInDir(te.gitRepoDir, "git", "checkout", "main")

	return te.gitRepoDir
}

// RunVendattaCommand runs a vendatta command in the test repo
func (te *TestEnvironment) RunVendattaCommand(args ...string) (string, error) {
	cmd := exec.Command(te.vendattaBin, args...)
	cmd.Dir = te.gitRepoDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunVendattaCommandAsync runs a vendatta command asynchronously
func (te *TestEnvironment) RunVendattaCommandAsync(args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(te.vendattaBin, args...)
	cmd.Dir = te.gitRepoDir
	return cmd, cmd.Start()
}

// GetSessions returns current vendatta sessions
func (te *TestEnvironment) GetSessions() ([]string, error) {
	output, err := te.RunVendattaCommand("list")
	if err != nil {
		return nil, err
	}

	var sessions []string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				sessions = append(sessions, parts[0])
			}
		}
	}
	return sessions, nil
}

// TestVendattaInit tests the vendatta init command
func TestVendattaInit(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.CleanupAllSessions()

	repoDir := te.SetupTestRepo()

	originalDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(originalDir)
	_, err := te.RunVendattaCommand("init")
	if err != nil {
		t.Fatalf("vendatta init failed: %v", err)
	}

	// Start dev session asynchronously (since it runs in background)
	cmd, err := te.RunVendattaCommandAsync("dev", "feature-1")
	if err != nil {
		t.Fatalf("vendatta dev failed to start: %v", err)
	}
	defer cmd.Process.Kill()

	// Wait for session to initialize
	time.Sleep(5 * time.Second)

	// Check that sessions exist
	sessions, err := te.GetSessions()
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	if len(sessions) == 0 {
		t.Fatal("No sessions found after vendatta dev")
	}

	// Verify worktree was created
	worktreePath := ".vendatta/worktrees/feature-1"
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Fatalf("Worktree directory %s was not created", worktreePath)
	}
}

// TestVendattaServiceDiscovery tests that environment variables are set correctly
func TestVendattaServiceDiscovery(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.CleanupAllSessions()

	repoDir := te.SetupTestRepo()

	originalDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(originalDir)

	_, err := te.RunVendattaCommand("init")
	if err != nil {
		t.Fatalf("vendatta init failed: %v", err)
	}
	configContent := `name: test-project
provider: docker
services:
  web:
    port: 3000
  api:
    port: 5000
docker:
  image: ubuntu:22.04
  dind: true
`
	if err := os.WriteFile(".vendatta/config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	cmd, err := te.RunVendattaCommandAsync("dev", "feature-1")
	if err != nil {
		t.Fatalf("vendatta dev failed to start: %v", err)
	}
	defer cmd.Process.Kill()

	time.Sleep(15 * time.Second)
	worktreeDir := ".vendatta/worktrees/feature-1"
	envFilePath := filepath.Join(worktreeDir, ".env")
	envContent, err := os.ReadFile(envFilePath)
	if err != nil {
		t.Fatalf("Failed to read .env file: %v", err)
	}

	envStr := string(envContent)
	if !strings.Contains(envStr, "OURSKY_SERVICE_WEB_URL=http://localhost:3000") {
		t.Fatal("OURSKY_SERVICE_WEB_URL not found in .env file")
	}
	if !strings.Contains(envStr, "OURSKY_SERVICE_API_URL=http://localhost:5000") {
		t.Fatal("OURSKY_SERVICE_API_URL not found in .env file")
	}
}

// TestVendattaSessionManagement tests session listing and killing
func TestVendattaSessionManagement(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.CleanupAllSessions()

	repoDir := te.SetupTestRepo()

	originalDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(originalDir)

	_, err := te.RunVendattaCommand("init")
	if err != nil {
		t.Fatalf("vendatta init failed: %v", err)
	}
	cmd1, err := te.RunVendattaCommandAsync("dev", "feature-1")
	if err != nil {
		t.Fatalf("Failed to start first session: %v", err)
	}
	defer cmd1.Process.Kill()

	cmd2, err := te.RunVendattaCommandAsync("dev", "feature-2")
	if err != nil {
		t.Fatalf("Failed to start second session: %v", err)
	}
	defer cmd2.Process.Kill()

	time.Sleep(15 * time.Second)
	sessions, err := te.GetSessions()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) < 2 {
		t.Fatalf("Expected at least 2 sessions, got %d", len(sessions))
	}

	// Kill first session
	if err := te.CleanupSession(sessions[0]); err != nil {
		t.Fatalf("Failed to kill session: %v", err)
	}

	// Verify session was killed
	time.Sleep(3 * time.Second)
	remainingSessions, err := te.GetSessions()
	if err != nil {
		t.Fatalf("Failed to list remaining sessions: %v", err)
	}

	if len(remainingSessions) >= len(sessions) {
		t.Fatal("Session was not killed successfully")
	}
}

// TestVendattaWorktreeIsolation tests that each branch gets its own isolated environment
func TestVendattaWorktreeIsolation(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.CleanupAllSessions()

	repoDir := te.SetupTestRepo()

	originalDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(originalDir)

	_, err := te.RunVendattaCommand("init")
	if err != nil {
		t.Fatalf("vendatta init failed: %v", err)
	}
	cmd1, err := te.RunVendattaCommandAsync("dev", "feature-1")
	if err != nil {
		t.Fatalf("Failed to start feature-1 session: %v", err)
	}
	defer cmd1.Process.Kill()

	cmd2, err := te.RunVendattaCommandAsync("dev", "feature-2")
	if err != nil {
		t.Fatalf("Failed to start feature-2 session: %v", err)
	}
	defer cmd2.Process.Kill()

	time.Sleep(15 * time.Second)
	wt1Path := ".vendatta/worktrees/feature-1"
	wt2Path := ".vendatta/worktrees/feature-2"

	if _, err := os.Stat(wt1Path); os.IsNotExist(err) {
		t.Fatalf("Worktree %s was not created", wt1Path)
	}
	if _, err := os.Stat(wt2Path); os.IsNotExist(err) {
		t.Fatalf("Worktree %s was not created", wt2Path)
	}
	testFile1 := filepath.Join(wt1Path, "unique-to-feature1.txt")
	testFile2 := filepath.Join(wt2Path, "unique-to-feature2.txt")

	if err := os.WriteFile(testFile1, []byte("feature1 content"), 0644); err != nil {
		t.Fatalf("Failed to create file in feature-1 worktree: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("feature2 content"), 0644); err != nil {
		t.Fatalf("Failed to create file in feature-2 worktree: %v", err)
	}
	content1, err := os.ReadFile(testFile1)
	if err != nil || string(content1) != "feature1 content" {
		t.Fatal("File in feature-1 worktree has incorrect content")
	}

	content2, err := os.ReadFile(testFile2)
	if err != nil || string(content2) != "feature2 content" {
		t.Fatal("File in feature-2 worktree has incorrect content")
	}
}

// CleanupSession cleans up a specific session
func (te *TestEnvironment) CleanupSession(sessionID string) error {
	_, err := te.RunVendattaCommand("kill", sessionID)
	return err
}

// CleanupAllSessions cleans up all active sessions
func (te *TestEnvironment) CleanupAllSessions() {
	sessions, err := te.GetSessions()
	if err != nil {
		te.t.Logf("Error getting sessions for cleanup: %v", err)
		return
	}

	for _, session := range sessions {
		if err := te.CleanupSession(session); err != nil {
			te.t.Logf("Error cleaning up session %s: %v", session, err)
		}
	}
}

// Helper methods
func (te *TestEnvironment) runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		te.t.Fatalf("Command failed: %s %v\nOutput: %s", name, args, output)
	}
}

func (te *TestEnvironment) runCommandInDir(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		te.t.Fatalf("Command failed in %s: %s %v\nOutput: %s", dir, name, args, output)
	}
}

func (te *TestEnvironment) createFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		te.t.Fatalf("Failed to create file %s: %v", path, err)
	}
}

func buildVendatta(t *testing.T, outputPath string) {
	// Get the project root (assuming we're in the test directory)
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")

	cmd := exec.Command("go", "build", "-o", outputPath, "cmd/oursky/main.go")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build vendatta binary: %v\nOutput: %s", err, output)
	}
}

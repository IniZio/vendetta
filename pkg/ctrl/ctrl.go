package ctrl

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vibegear/oursky/pkg/config"
	"github.com/vibegear/oursky/pkg/provider"
	"github.com/vibegear/oursky/pkg/worktree"
)

type Controller interface {
	Init(ctx context.Context) error
	Dev(ctx context.Context, branch string) error
	WorkspaceCreate(ctx context.Context, name string) error
	WorkspaceUp(ctx context.Context, name string) error
	WorkspaceDown(ctx context.Context, name string) error
	WorkspaceShell(ctx context.Context, name string) error
	WorkspaceList(ctx context.Context) error
	WorkspaceRm(ctx context.Context, name string) error
	Kill(ctx context.Context, sessionID string) error
	List(ctx context.Context) ([]provider.Session, error)
	Exec(ctx context.Context, sessionID string, cmd []string) error
}

type BaseController struct {
	Providers       map[string]provider.Provider
	WorktreeManager worktree.Manager
}

func NewBaseController(providers []provider.Provider, wtManager worktree.Manager) *BaseController {
	pMap := make(map[string]provider.Provider)
	for _, p := range providers {
		pMap[p.Name()] = p
	}
	if wtManager == nil {
		wtManager = worktree.NewManager(".", ".vendatta/worktrees")
	}
	return &BaseController{
		Providers:       pMap,
		WorktreeManager: wtManager,
	}
}

func (c *BaseController) Init(ctx context.Context) error {
	dirs := []string{
		".vendatta/hooks",
		".vendatta/worktrees",
		".vendatta/agents/rules",
		".vendatta/agents/skills",
		".vendatta/agents/commands",
		".vendatta/templates/skills",
		".vendatta/templates/rules",
		".vendatta/templates/commands",
		".vendatta/remotes",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	configYaml := `name: my-project
provider: docker
services:
  web:
    port: 3000
docker:
  image: ubuntu:22.04
  dind: true
hooks:
  up: .vendatta/hooks/up.sh
`
	if err := os.WriteFile(".vendatta/config.yaml", []byte(configYaml), 0644); err != nil {
		return err
	}

	upSh := `#!/bin/bash
# Main startup script - replace this with your development workflow
echo "Starting development environment..."

# Install dependencies if needed
if ! command -v node &> /dev/null; then
    echo "Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
    apt-get install -y nodejs
fi

# Install docker-compose if not present
if ! command -v docker-compose &> /dev/null; then
    echo "Installing docker-compose..."
    apt-get update && apt-get install -y docker-compose
fi

# Start your services here
echo "Starting services..."

# Example: Start database
docker-compose up -d postgres &

# Example: Start API server
cd /workspace/server && npm install && HOST=0.0.0.0 PORT=5000 npm run dev &
API_PID=$!

sleep 5

# Example: Start web client
cd /workspace/client && npm install && HOST=0.0.0.0 PORT=3000 npm run dev &
WEB_PID=$!

echo "Services starting... PIDs: API($API_PID), WEB($WEB_PID)"
echo "Development environment ready."

# Keep container alive
wait
`
	if err := os.WriteFile(".vendatta/hooks/up.sh", []byte(upSh), 0755); err != nil {
		return err
	}

	baseRule := `# Base Rules
- Follow existing code patterns.
- Ensure type safety.
`
	if err := os.WriteFile(".vendatta/agents/rules/base.md", []byte(baseRule), 0644); err != nil {
		return err
	}

	agentDirs := []string{
		".vendatta/agents/cursor",
		".vendatta/agents/opencode",
		".vendatta/agents/claude-desktop",
		".vendatta/agents/claude-code",
	}
	for _, dir := range agentDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	cursorTpl := `{
  "mcpServers": {
    "oursky": {
      "command": "vendatta",
      "args": ["agent", "{{.ProjectName}}"],
      "port": {{.Port}}
    }
  }
}
`
	if err := os.WriteFile(".vendatta/agents/cursor/mcp.json.tpl", []byte(cursorTpl), 0644); err != nil {
		return err
	}

	opencodeTpl := `{
  "name": "{{.ProjectName}}",
  "mcp": {
    "port": {{.Port}}
  }
}
`
	if err := os.WriteFile(".vendatta/agents/opencode/opencode.json.tpl", []byte(opencodeTpl), 0644); err != nil {
		return err
	}

	return nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip worktrees directory to avoid infinite recursion
		if entry.Name() == "worktrees" {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

func (c *BaseController) WorkspaceCreate(ctx context.Context, name string) error {
	fmt.Printf("🚀 Creating workspace '%s'...\n", name)

	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize remotes
	fmt.Println("📦 Initializing template remotes...")
	if err := cfg.InitializeRemotes(".vendatta"); err != nil {
		return fmt.Errorf("failed to initialize remotes: %w", err)
	}

	// Get merged templates
	fmt.Println("🔧 Merging AI agent templates...")
	merged, err := cfg.GetMergedTemplates(".vendatta")
	if err != nil {
		return fmt.Errorf("failed to merge templates: %w", err)
	}

	fmt.Println("🌳 Setting up Git worktree...")
	wtPath, err := c.WorktreeManager.Add(name)
	if err != nil {
		return fmt.Errorf("failed to setup worktree: %w", err)
	}

	// Copy .vendatta config to worktree so hooks can execute
	if err := copyDir(".vendatta", filepath.Join(wtPath, ".vendatta")); err != nil {
		return fmt.Errorf("failed to copy vendatta config to worktree: %w", err)
	}

	absWtPath, err := filepath.Abs(wtPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for worktree: %w", err)
	}

	// Generate agent configs
	fmt.Println("🤖 Generating AI agent configurations...")
	if err := cfg.GenerateAgentConfigs(absWtPath, merged); err != nil {
		return fmt.Errorf("failed to generate agent configs: %w", err)
	}
	fmt.Printf("📂 Worktree: %s\n", absWtPath)
	fmt.Println("💡 Open this directory in your AI agent (Cursor, OpenCode, etc.)")
	fmt.Println("🔍 Use 'vendatta list' to see active sessions")
	return nil
}

func (c *BaseController) WorkspaceUp(ctx context.Context, name string) error {
	// Find project root (where .vendatta directory is)
	projectRoot, err := c.findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Auto-detect workspace if name is empty
	if name == "" {
		var err error
		name, err = c.detectWorkspaceFromCWD()
		if err != nil {
			return fmt.Errorf("no workspace specified and auto-detection failed: %w", err)
		}
		fmt.Printf("📍 Auto-detected workspace: %s\n", name)
	}

	fmt.Printf("🚀 Starting workspace '%s'...\n", name)

	// Change to project root for config loading
	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(projectRoot); err != nil {
		return fmt.Errorf("failed to change to project root: %w", err)
	}

	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	wtPath := filepath.Join(".vendatta", "worktrees", name)
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("workspace '%s' does not exist (worktree not found at %s)", name, wtPath)
	}

	absWtPath, err := filepath.Abs(wtPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for worktree: %w", err)
	}

	if cfg.Provider != "" {
		p, ok := c.Providers[cfg.Provider]
		if !ok {
			return fmt.Errorf("provider '%s' not found", cfg.Provider)
		}

		sessionID := fmt.Sprintf("%s-%s", cfg.Name, name)
		fmt.Printf("🐳 Creating %s session...\n", cfg.Provider)
		session, err := p.Create(ctx, sessionID, absWtPath, cfg)
		if err != nil {
			sessions, _ := p.List(ctx)
			for _, s := range sessions {
				if s.ID == sessionID || s.Labels["oursky.session.id"] == sessionID {
					session = &s
					break
				}
			}
			if session == nil {
				return fmt.Errorf("failed to create session: %w", err)
			}
		}

		fmt.Println("▶️  Starting session...")
		if err := p.Start(ctx, session.ID); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}

		if err := c.setupWorkspaceEnvironment(ctx, session, cfg, p, absWtPath); err != nil {
			return fmt.Errorf("failed to setup workspace environment: %w", err)
		}
	}

	upHookPath := filepath.Join(absWtPath, ".vendatta", "hooks", "up.sh")
	if _, err := os.Stat(upHookPath); err == nil {
		fmt.Println("🔧 Running up hook (local startup)...")
		if err := c.runHook(ctx, upHookPath, cfg, absWtPath); err != nil {
			return fmt.Errorf("up hook failed: %w", err)
		}
		fmt.Println("✅ Up hook completed successfully")
	} else {
		fmt.Println("ℹ️  No up.sh hook found")
	}

	fmt.Printf("\n🎉 Workspace '%s' is ready!\n", name)
	fmt.Printf("📂 Worktree: %s\n", absWtPath)
	fmt.Printf("🛑 Run 'vendatta workspace down %s' to stop\n", name)
	return nil
}

func (c *BaseController) handleBranchConflicts(branchName string) error {
	// Check if branch exists
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	cmd.Dir = "."
	if cmd.Run() == nil {
		// Branch exists, check for uncommitted changes
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusCmd.Dir = "."
		output, err := statusCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}

		if len(output) > 0 {
			fmt.Println("📦 Stashing uncommitted changes...")
			stashCmd := exec.Command("git", "stash", "push", "-m", "vendatta: auto-stash before workspace creation")
			stashCmd.Dir = "."
			if err := stashCmd.Run(); err != nil {
				return fmt.Errorf("failed to stash changes: %w", err)
			}
			fmt.Println("✅ Changes stashed successfully")
		}
	}
	return nil
}

func (c *BaseController) runHook(ctx context.Context, hookPath string, cfg *config.Config, workspacePath string) error {
	// Make hook executable
	if err := os.Chmod(hookPath, 0755); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	// Prepare environment variables
	env := []string{
		fmt.Sprintf("WORKSPACE_NAME=%s", filepath.Base(workspacePath)),
		fmt.Sprintf("WORKTREE_PATH=%s", workspacePath),
	}

	// Add service discovery variables
	envFileContent := ""
	for name, svc := range cfg.Services {
		var port int
		var url string

		if svc.Port > 0 {
			// Legacy port configuration
			port = svc.Port
		} else {
			// Auto-detect port from command
			port = detectPortFromCommand(svc.Command)
		}

		if port > 0 {
			// Protocol detection
			protocol := detectProtocol(name, svc.Command)
			url = fmt.Sprintf("%s://localhost:%d", protocol, port)

			envVar := fmt.Sprintf("OURSKY_SERVICE_%s_URL=%s", strings.ToUpper(name), url)
			env = append(env, envVar)
			envFileContent += envVar + "\n"
		}
	}

	// Write environment variables to .env file
	if envFileContent != "" {
		envFilePath := filepath.Join(workspacePath, ".env")
		if err := os.WriteFile(envFilePath, []byte(envFileContent), 0644); err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
	}

	// Run the hook
	cmd := exec.CommandContext(ctx, "/bin/bash", hookPath)
	cmd.Dir = workspacePath
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *BaseController) findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for .vendatta directory
	// Skip .vendatta directories that are inside worktrees (copies)
	currentDir := cwd
	for {
		vendattaPath := filepath.Join(currentDir, ".vendatta")
		if _, err := os.Stat(vendattaPath); err == nil {
			// Check if this .vendatta is inside a worktrees directory (indicating a copy)
			if !strings.Contains(currentDir, "/worktrees/") && !strings.Contains(currentDir, "\\worktrees\\") {
				return currentDir, nil
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("could not find .vendatta directory (not in a vendatta project)")
}

func (c *BaseController) detectWorkspaceFromCWD() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for .vendatta/worktrees
	currentDir := cwd
	for {
		worktreesPath := filepath.Join(currentDir, ".vendatta", "worktrees")
		if _, err := os.Stat(worktreesPath); err == nil {
			// Found worktrees directory, check if cwd is inside it
			relPath, err := filepath.Rel(worktreesPath, cwd)
			if err != nil {
				break
			}

			if !strings.HasPrefix(relPath, "..") && relPath != "." {
				// Extract workspace name from path
				parts := strings.Split(relPath, string(filepath.Separator))
				if len(parts) > 0 && parts[0] != "" {
					return parts[0], nil
				}
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("not in a vendatta worktree")
}

func (c *BaseController) setupWorkspaceEnvironment(ctx context.Context, session *provider.Session, cfg *config.Config, p provider.Provider, absWtPath string) error {
	sessions, _ := p.List(ctx)
	var activeSession *provider.Session
	for _, s := range sessions {
		if s.ID == session.ID || s.Labels["oursky.session.id"] == session.ID {
			activeSession = &s
			break
		}
	}

	env := []string{}
	if activeSession != nil {
		fmt.Println("🌐 Service port mappings:")
		for name, svc := range cfg.Services {
			if svc.Port > 0 {
				pStr := fmt.Sprintf("%d", svc.Port)
				if publicPort, ok := activeSession.Services[pStr]; ok {
					url := fmt.Sprintf("http://localhost:%d", publicPort)
					envVar := fmt.Sprintf("OURSKY_SERVICE_%s_URL=%s", strings.ToUpper(name), url)
					env = append(env, envVar)
					fmt.Printf("  📍 %s → %s\n", strings.ToUpper(name), url)
				}
			}
		}
	}

	// Create .env file in worktree with service URLs
	if len(env) > 0 {
		envFilePath := filepath.Join(absWtPath, ".env")
		envContent := strings.Join(env, "\n") + "\n"
		if err := os.WriteFile(envFilePath, []byte(envContent), 0644); err != nil {
			return fmt.Errorf("failed to create .env file: %w", err)
		}
		fmt.Printf("📄 Created .env file with service URLs\n")
	}

	if cfg.Hooks.Setup != "" {
		fmt.Printf("🔧 Running setup hook: %s\n", cfg.Hooks.Setup)
		err := p.Exec(ctx, session.ID, provider.ExecOptions{
			Cmd:    []string{"/bin/bash", "/workspace/" + cfg.Hooks.Setup},
			Env:    env,
			Stdout: true,
		})
		if err != nil {
			return fmt.Errorf("setup hook failed: %w", err)
		}
		fmt.Println("✅ Setup hook completed successfully")
	}

	return nil
}

func (c *BaseController) WorkspaceDown(ctx context.Context, name string) error {
	// Auto-detect workspace if name is empty
	if name == "" {
		detectedName, err := c.detectWorkspaceFromCWD()
		if err != nil {
			return fmt.Errorf("no workspace specified and auto-detection failed: %w", err)
		}
		name = detectedName
		fmt.Printf("📍 Auto-detected workspace: %s\n", name)
	}

	fmt.Printf("🛑 Stopping workspace '%s'...\n", name)

	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	sessionID := fmt.Sprintf("%s-%s", cfg.Name, name)

	for _, p := range c.Providers {
		sessions, _ := p.List(ctx)
		for _, s := range sessions {
			if s.ID == sessionID || s.Labels["oursky.session.id"] == sessionID {
				fmt.Printf("🧹 Destroying session %s...\n", s.ID)
				return p.Destroy(ctx, s.ID)
			}
		}
	}

	return fmt.Errorf("workspace '%s' not found", name)
}

func (c *BaseController) WorkspaceShell(ctx context.Context, name string) error {
	// Auto-detect workspace if name is empty
	if name == "" {
		detectedName, err := c.detectWorkspaceFromCWD()
		if err != nil {
			return fmt.Errorf("no workspace specified and auto-detection failed: %w", err)
		}
		name = detectedName
		fmt.Printf("📍 Auto-detected workspace: %s\n", name)
	}

	fmt.Printf("🐚 Opening shell in workspace '%s'...\n", name)

	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	sessionID := fmt.Sprintf("%s-%s", cfg.Name, name)

	for _, p := range c.Providers {
		sessions, _ := p.List(ctx)
		for _, s := range sessions {
			if s.ID == sessionID || s.Labels["oursky.session.id"] == sessionID {
				return p.Exec(ctx, s.ID, provider.ExecOptions{
					Cmd:    []string{"/bin/bash"},
					Stdout: true,
					Stderr: true,
				})
			}
		}
	}

	return fmt.Errorf("workspace '%s' not running", name)
}

func (c *BaseController) WorkspaceList(ctx context.Context) error {
	fmt.Println("📋 Active workspaces:")

	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	activeWorkspaces := make(map[string]provider.Session)
	for _, p := range c.Providers {
		sessions, err := p.List(ctx)
		if err != nil {
			continue
		}

		for _, s := range sessions {
			if sessionID, ok := s.Labels["oursky.session.id"]; ok {
				if strings.HasPrefix(sessionID, cfg.Name+"-") {
					workspaceName := strings.TrimPrefix(sessionID, cfg.Name+"-")
					activeWorkspaces[workspaceName] = s
				}
			}
		}
	}

	worktreesDir := ".vendatta/worktrees"
	entries, _ := os.ReadDir(worktreesDir)
	hasWorkspaces := false

	allNames := make(map[string]bool)
	for name := range activeWorkspaces {
		allNames[name] = true
	}
	for _, entry := range entries {
		if entry.IsDir() {
			allNames[entry.Name()] = true
		}
	}

	for name := range allNames {
		status := "created"
		providerName := "-"
		if s, ok := activeWorkspaces[name]; ok {
			status = s.Status
			providerName = s.Provider
		}
		fmt.Printf("  %s (%s) - %s\n", name, providerName, status)
		hasWorkspaces = true
	}

	if !hasWorkspaces {
		fmt.Println("  No active workspaces")
	}

	return nil
}

func (c *BaseController) WorkspaceRm(ctx context.Context, name string) error {
	fmt.Printf("🗑️  Removing workspace '%s'...\n", name)

	worktreePath := filepath.Join(".vendatta", "worktrees", name)
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace '%s' not found", name)
	}

	if err := c.WorkspaceDown(ctx, name); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to stop workspace: %w", err)
		}
	}
	if err := c.WorktreeManager.Remove(name); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	fmt.Printf("✅ Workspace '%s' removed successfully\n", name)
	return nil
}

func (c *BaseController) Kill(ctx context.Context, sessionID string) error {
	for _, p := range c.Providers {
		sessions, _ := p.List(ctx)
		for _, s := range sessions {
			if s.ID == sessionID || s.Labels["oursky.session.id"] == sessionID {
				return p.Destroy(ctx, s.ID)
			}
		}
	}
	return fmt.Errorf("session %s not found", sessionID)
}

func (c *BaseController) List(ctx context.Context) ([]provider.Session, error) {
	var all []provider.Session
	for _, p := range c.Providers {
		sessions, err := p.List(ctx)
		if err != nil {
			continue
		}
		all = append(all, sessions...)
	}
	return all, nil
}

func (c *BaseController) Exec(ctx context.Context, sessionID string, cmd []string) error {
	for _, p := range c.Providers {
		sessions, _ := p.List(ctx)
		for _, s := range sessions {
			if s.ID == sessionID || s.Labels["oursky.session.id"] == sessionID {
				return p.Exec(ctx, s.ID, provider.ExecOptions{
					Cmd:    cmd,
					Stdout: true,
				})
			}
		}
	}
	return fmt.Errorf("session %s not found", sessionID)
}

func detectPortFromCommand(command string) int {
	re := regexp.MustCompile(`PORT=(\d+)`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		var port int
		fmt.Sscanf(matches[1], "%d", &port)
		if port > 0 {
			return port
		}
	}

	portPatterns := []struct {
		pattern string
		port    int
	}{
		{"npm run dev", 3000},
		{"npm start", 3000},
		{"yarn dev", 3000},
		{"yarn start", 3000},
		{"python.*manage.py.*runserver", 8000},
		{"django.*runserver", 8000},
		{"flask run", 5000},
		{"rails server", 3000},
		{"rails s", 3000},
		{"docker-compose.*postgres", 5432},
		{"docker-compose.*mysql", 3306},
		{"docker-compose.*mongodb", 27017},
		{"docker-compose.*redis", 6379},
	}

	commandLower := strings.ToLower(command)
	for _, pp := range portPatterns {
		matched, _ := regexp.MatchString(strings.ToLower(pp.pattern), commandLower)
		if matched {
			return pp.port
		}
	}

	return 0 // No port detected
}

// detectProtocol determines the protocol based on service name and command
func detectProtocol(serviceName, command string) string {
	commandLower := strings.ToLower(command)

	// Database services
	if strings.Contains(commandLower, "postgres") || strings.Contains(commandLower, "postgresql") {
		return "postgresql"
	}
	if strings.Contains(commandLower, "mysql") {
		return "mysql"
	}
	if strings.Contains(commandLower, "mongodb") {
		return "mongodb"
	}
	if strings.Contains(commandLower, "redis") {
		return "redis"
	}

	// Default to http for web services
	return "http"
}

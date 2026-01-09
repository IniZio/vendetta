package ctrl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vibegear/oursky/pkg/config"
	"github.com/vibegear/oursky/pkg/provider"
	"github.com/vibegear/oursky/pkg/worktree"
)

type Controller interface {
	Init(ctx context.Context) error
	Dev(ctx context.Context, branch string) error
	Kill(ctx context.Context, sessionID string) error
	List(ctx context.Context) ([]provider.Session, error)
	Exec(ctx context.Context, sessionID string, cmd []string) error
}

type BaseController struct {
	Providers map[string]provider.Provider
}

func NewBaseController(providers []provider.Provider) *BaseController {
	pMap := make(map[string]provider.Provider)
	for _, p := range providers {
		pMap[p.Name()] = p
	}
	return &BaseController{Providers: pMap}
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
  setup: .vendatta/hooks/setup.sh
  dev: .vendatta/hooks/dev.sh
`
	if err := os.WriteFile(".vendatta/config.yaml", []byte(configYaml), 0644); err != nil {
		return err
	}

	setupSh := `#!/bin/bash
# Install docker if dind is enabled
echo "Setting up environment..."
`
	if err := os.WriteFile(".vendatta/hooks/setup.sh", []byte(setupSh), 0755); err != nil {
		return err
	}

	baseRule := `# Base Rules
- Follow existing code patterns.
- Ensure type safety.
`
	if err := os.WriteFile(".vendatta/agents/rules/base.md", []byte(baseRule), 0644); err != nil {
		return err
	}

	return nil
}

func (c *BaseController) Dev(ctx context.Context, branch string) error {
	fmt.Printf("🚀 Starting dev session for branch '%s'...\n", branch)

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

	p, ok := c.Providers[cfg.Provider]
	if !ok {
		return fmt.Errorf("provider %s not found", cfg.Provider)
	}

	fmt.Println("🌳 Setting up Git worktree...")
	wtManager := worktree.NewManager(".", ".vendatta/worktrees")
	wtPath, err := wtManager.Add(branch)
	if err != nil {
		return fmt.Errorf("failed to setup worktree: %w", err)
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

	fmt.Printf("🐳 Creating %s session...\n", cfg.Provider)
	sessionID := fmt.Sprintf("%s-%s", cfg.Name, branch)

	existingSessions, _ := p.List(ctx)
	for _, s := range existingSessions {
		if s.ID == sessionID || s.Labels["oursky.session.id"] == sessionID {
			fmt.Printf("🧹 Cleaning up existing session %s...\n", s.ID)
			p.Destroy(ctx, s.ID)
			break
		}
	}

	session, err := p.Create(ctx, sessionID, absWtPath, cfg)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Println("▶️  Starting session...")
	if err := p.Start(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	sessions, _ := p.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Println("▶️  Starting session...")
	if err := p.Start(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	fmt.Println("▶️  Starting session...")
	if err := p.Start(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	var activeSession *provider.Session
	for _, s := range sessions {
		if s.ID == session.ID || s.Labels["oursky.session.id"] == sessionID {
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

	if cfg.Hooks.Setup != "" {
		fmt.Printf("🔧 Running setup hook: %s\n", cfg.Hooks.Setup)
		err = p.Exec(ctx, session.ID, provider.ExecOptions{
			Cmd:    []string{"/bin/bash", "/workspace/" + cfg.Hooks.Setup},
			Env:    env,
			Stdout: true,
		})
		if err != nil {
			return fmt.Errorf("setup hook failed: %w", err)
		}
		fmt.Println("✅ Setup hook completed successfully")
	}

	fmt.Printf("\n🎉 Session %s is ready!\n", session.ID)
	fmt.Printf("📂 Worktree: %s\n", absWtPath)
	fmt.Println("💡 Open this directory in your AI agent (Cursor, OpenCode, etc.)")
	fmt.Println("🔍 Use 'vendatta list' to see active sessions")
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

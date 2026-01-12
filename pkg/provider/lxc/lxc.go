package lxc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vibegear/vendatta/pkg/provider"
)

type LXCProvider struct {
	name string
}

func NewLXCProvider() (provider.Provider, error) {
	// Check if lxc command is available
	if _, err := exec.LookPath("lxc"); err != nil {
		return nil, fmt.Errorf("lxc command not found: %w", err)
	}
	return &LXCProvider{name: "lxc"}, nil
}

func (p *LXCProvider) Name() string {
	return p.name
}

func (p *LXCProvider) Create(ctx context.Context, sessionID string, workspacePath string, config interface{}) (*provider.Session, error) {
	containerName := fmt.Sprintf("vendatta-%s", sessionID)

	cmd := exec.CommandContext(ctx, "lxc", "init", "ubuntu:22.04", containerName, "--config", "limits.memory=512MB")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to init LXC container: %w: %s", err, string(output))
	}

	// Add bind mount for worktree
	mountCmd := exec.CommandContext(ctx, "lxc", "config", "device", "add", containerName, "worktree", "disk", fmt.Sprintf("source=%s", workspacePath), "path=/workspace")
	mountOutput, mountErr := mountCmd.CombinedOutput()
	if mountErr != nil {
		p.Destroy(ctx, sessionID)
		return nil, fmt.Errorf("failed to mount worktree: %w: %s", mountErr, string(mountOutput))
	}

	session := &provider.Session{
		ID:       sessionID,
		Provider: p.name,
		Status:   "",
		Labels: map[string]string{
			"vendatta.session.id": sessionID,
		},
	}

	return session, nil
}

func (p *LXCProvider) Start(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("vendatta-%s", sessionID)
	cmd := exec.CommandContext(ctx, "lxc", "start", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

func (p *LXCProvider) Stop(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("vendatta-%s", sessionID)
	cmd := exec.CommandContext(ctx, "lxc", "stop", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

func (p *LXCProvider) Destroy(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("vendatta-%s", sessionID)

	// lxc delete will implicitly stop the container
	cmd := exec.CommandContext(ctx, "lxc", "delete", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

func (p *LXCProvider) List(ctx context.Context) ([]provider.Session, error) {
	cmd := exec.CommandContext(ctx, "lxc", "list", "--format", "csv", "-c", "n,s")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list LXC containers: %w", err)
	}

	var sessions []provider.Session
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 2 && strings.HasPrefix(parts[0], "vendatta-") {
			sessionID := strings.TrimPrefix(parts[0], "vendatta-")
			status := strings.ToLower(parts[1])

			sessions = append(sessions, provider.Session{
				ID:       sessionID,
				Provider: p.name,
				Status:   status,
				Labels: map[string]string{
					"vendatta.session.id": sessionID,
				},
			})
		}
	}

	return sessions, nil
}

func (p *LXCProvider) Exec(ctx context.Context, sessionID string, opts provider.ExecOptions) error {
	containerName := fmt.Sprintf("vendatta-%s", sessionID)

	args := []string{"exec", containerName, "--"}
	args = append(args, opts.Cmd...)

	cmd := exec.CommandContext(ctx, "lxc", args...)
	cmd.Env = append(os.Environ(), opts.Env...)

	if opts.Stdout {
		if opts.StdoutWriter != nil {
			cmd.Stdout = opts.StdoutWriter
		}
	}
	if opts.Stderr {
		if opts.StderrWriter != nil {
			cmd.Stderr = opts.StderrWriter
		}
	}

	if opts.StdoutWriter != nil || opts.StderrWriter != nil {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute command in LXC container %s: %w", containerName, err)
		}
		return nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command in LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

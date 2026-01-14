package lxc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vibegear/vendetta/pkg/config"
	"github.com/vibegear/vendetta/pkg/provider"
	"github.com/vibegear/vendetta/pkg/transport"
)

type LXCProvider struct {
	name string
	transport.Manager
	remote string
}

func NewLXCProvider() (provider.Provider, error) {
	// Check if lxc command is available
	if _, err := exec.LookPath("lxc"); err != nil {
		return nil, fmt.Errorf("lxc command not found: %w", err)
	}
	return &LXCProvider{
		name:    "lxc",
		Manager: *transport.NewManager(),
	}, nil
}

func (p *LXCProvider) Name() string {
	return p.name
}

func (p *LXCProvider) Create(ctx context.Context, sessionID string, workspacePath string, rawConfig interface{}) (*provider.Session, error) {
	var cfg *config.Config
	if rawConfig != nil {
		var ok bool
		cfg, ok = rawConfig.(*config.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type")
		}
	} else {
		cfg = &config.Config{}
	}

	if cfg.Remote.Node != "" {
		p.remote = fmt.Sprintf("%s@%s", cfg.Remote.User, cfg.Remote.Node)
		if cfg.Remote.Port > 0 && cfg.Remote.Port != 22 {
			p.remote = fmt.Sprintf("%s -p %d", p.remote, cfg.Remote.Port)
		}

		target := cfg.Remote.Node
		if cfg.Remote.Port > 0 && cfg.Remote.Port != 22 {
			target = fmt.Sprintf("%s:%d", cfg.Remote.Node, cfg.Remote.Port)
		} else {
			target = fmt.Sprintf("%s:22", cfg.Remote.Node)
		}

		sshKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")

		sshConfig := transport.CreateDefaultSSHConfig(
			target,
			cfg.Remote.User,
			sshKeyPath,
		)
		if err := p.RegisterConfig("remote-lxc", sshConfig); err != nil {
			return nil, fmt.Errorf("failed to register SSH transport config: %w", err)
		}

		return p.createRemote(ctx, sessionID, workspacePath, cfg)
	}

	return p.createLocal(ctx, sessionID, workspacePath, cfg)
}

func (p *LXCProvider) Start(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("vendetta-%s", sessionID)

	if p.remote != "" {
		t, err := p.CreateTransport("remote-lxc")
		if err != nil {
			return fmt.Errorf("failed to create SSH transport: %w", err)
		}

		err = t.Connect(ctx, p.remote)
		if err != nil {
			return fmt.Errorf("failed to connect via transport: %w", err)
		}
		defer t.Disconnect(ctx)

		result, err := t.Execute(ctx, &transport.Command{
			Cmd:           []string{"lxc", "start", containerName},
			CaptureOutput: true,
		})
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("lxc start failed: %s", result.Output)
		}
		return nil
	}

	cmd := exec.CommandContext(ctx, "lxc", "start", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

func (p *LXCProvider) Stop(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("vendetta-%s", sessionID)

	if p.remote != "" {
		t, err := p.CreateTransport("remote-lxc")
		if err != nil {
			return fmt.Errorf("failed to create SSH transport: %w", err)
		}

		err = t.Connect(ctx, p.remote)
		if err != nil {
			return fmt.Errorf("failed to connect via transport: %w", err)
		}
		defer t.Disconnect(ctx)

		result, err := t.Execute(ctx, &transport.Command{
			Cmd:           []string{"lxc", "stop", containerName},
			CaptureOutput: true,
		})
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("lxc stop failed: %s", result.Output)
		}
		return nil
	}

	cmd := exec.CommandContext(ctx, "lxc", "stop", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

func (p *LXCProvider) Destroy(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("vendetta-%s", sessionID)

	if p.remote != "" {
		t, err := p.CreateTransport("remote-lxc")
		if err != nil {
			return fmt.Errorf("failed to create SSH transport: %w", err)
		}

		err = t.Connect(ctx, p.remote)
		if err != nil {
			return fmt.Errorf("failed to connect via transport: %w", err)
		}
		defer t.Disconnect(ctx)

		result, err := t.Execute(ctx, &transport.Command{
			Cmd:           []string{"lxc", "delete", containerName},
			CaptureOutput: true,
		})
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("lxc delete failed: %s", result.Output)
		}
		return nil
	}

	cmd := exec.CommandContext(ctx, "lxc", "delete", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete LXC container %s: %w: %s", containerName, err, string(output))
	}
	return nil
}

func (p *LXCProvider) List(ctx context.Context) ([]provider.Session, error) {
	if p.remote != "" {
		return p.listRemote(ctx)
	}
	return p.listLocal(ctx)
}

func (p *LXCProvider) Exec(ctx context.Context, sessionID string, opts provider.ExecOptions) error {
	if p.remote != "" {
		containerName := fmt.Sprintf("vendetta-%s", sessionID)
		lxcCmd := append([]string{"lxc", "exec", containerName, "--"}, opts.Cmd...)

		t, err := p.CreateTransport("remote-lxc")
		if err != nil {
			return fmt.Errorf("failed to create SSH transport: %w", err)
		}

		err = t.Connect(ctx, p.remote)
		if err != nil {
			return fmt.Errorf("failed to connect via transport: %w", err)
		}
		defer t.Disconnect(ctx)

		envMap := make(map[string]string)
		for _, env := range opts.Env {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		result, err := t.Execute(ctx, &transport.Command{
			Cmd:           lxcCmd,
			Env:           envMap,
			CaptureOutput: false,
			Stdout:        opts.StdoutWriter,
			Stderr:        opts.StderrWriter,
		})
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("lxc exec failed with exit code %d", result.ExitCode)
		}
		return nil
	}

	containerName := fmt.Sprintf("vendetta-%s", sessionID)

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

func (p *LXCProvider) createLocal(ctx context.Context, sessionID string, workspacePath string, cfg *config.Config) (*provider.Session, error) {
	containerName := fmt.Sprintf("vendetta-%s", sessionID)

	cmd := exec.CommandContext(ctx, "lxc", "init", "ubuntu:22.04", containerName, "--config", "limits.memory=512MB")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to init LXC container: %w: %s", err, string(output))
	}

	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		_ = p.Destroy(ctx, sessionID)
		return nil, fmt.Errorf("failed to create workspace path: %w", err)
	}

	mountCmd := exec.CommandContext(ctx, "lxc", "config", "device", "add", containerName, "worktree", "disk", fmt.Sprintf("source=%s", workspacePath), "path=/workspace")
	mountOutput, mountErr := mountCmd.CombinedOutput()
	if mountErr != nil {
		_ = p.Destroy(ctx, sessionID)
		return nil, fmt.Errorf("failed to mount worktree: %w: %s", mountErr, string(mountOutput))
	}

	session := &provider.Session{
		ID:       sessionID,
		Provider: p.name,
		Status:   "",
		Labels: map[string]string{
			"vendetta.session.id": sessionID,
		},
	}

	return session, nil
}

func (p *LXCProvider) createRemote(ctx context.Context, sessionID string, workspacePath string, cfg *config.Config) (*provider.Session, error) {
	containerName := fmt.Sprintf("vendetta-%s", sessionID)

	t, err := p.CreateTransport("remote-lxc")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH transport: %w", err)
	}

	err = t.Connect(ctx, p.remote)
	if err != nil {
		return nil, fmt.Errorf("failed to connect via transport: %w", err)
	}
	defer t.Disconnect(ctx)

	result, err := t.Execute(ctx, &transport.Command{
		Cmd:           []string{"lxc", "init", "ubuntu:22.04", containerName, "--config", "limits.memory=512MB"},
		CaptureOutput: true,
	})
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("lxc init failed: %s", result.Output)
	}

	if _, err := t.Execute(ctx, &transport.Command{
		Cmd:           []string{"mkdir", "-p", workspacePath},
		CaptureOutput: true,
	}); err != nil {
		return nil, err
	}

	mountResult, err := t.Execute(ctx, &transport.Command{
		Cmd:           []string{"lxc", "config", "device", "add", containerName, "worktree", "disk", fmt.Sprintf("source=%s", workspacePath), "path=/workspace"},
		CaptureOutput: true,
	})
	if err != nil {
		return nil, err
	}
	if mountResult.ExitCode != 0 {
		return nil, fmt.Errorf("failed to mount worktree: %s", mountResult.Output)
	}

	return &provider.Session{
		ID:       sessionID,
		Provider: p.name,
		Status:   "",
		Labels: map[string]string{
			"vendetta.session.id": sessionID,
		},
	}, nil
}

func (p *LXCProvider) listLocal(ctx context.Context) ([]provider.Session, error) {
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
		if len(parts) >= 2 && strings.HasPrefix(parts[0], "vendetta-") {
			sessionID := strings.TrimPrefix(parts[0], "vendetta-")
			status := strings.ToLower(parts[1])

			sessions = append(sessions, provider.Session{
				ID:       sessionID,
				Provider: p.name,
				Status:   status,
				Labels: map[string]string{
					"vendetta.session.id": sessionID,
				},
			})
		}
	}

	return sessions, nil
}

func (p *LXCProvider) listRemote(ctx context.Context) ([]provider.Session, error) {
	t, err := p.CreateTransport("remote-lxc")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH transport: %w", err)
	}

	err = t.Connect(ctx, p.remote)
	if err != nil {
		return nil, fmt.Errorf("failed to connect via transport: %w", err)
	}
	defer t.Disconnect(ctx)

	result, err := t.Execute(ctx, &transport.Command{
		Cmd:           []string{"lxc", "list", "--format", "csv", "-c", "n,s"},
		CaptureOutput: true,
	})
	if err != nil {
		return nil, err
	}

	var sessions []provider.Session
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 2 && strings.HasPrefix(parts[0], "vendetta-") {
			sessionID := strings.TrimPrefix(parts[0], "vendetta-")
			status := strings.ToLower(parts[1])

			sessions = append(sessions, provider.Session{
				ID:       sessionID,
				Provider: p.name,
				Status:   status,
				Labels: map[string]string{
					"vendetta.session.id": sessionID,
				},
			})
		}
	}

	return sessions, nil
}

package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/vibegear/oursky/pkg/config"
	"github.com/vibegear/oursky/pkg/provider"
)

type DockerProvider struct {
	cli *client.Client
}

func NewDockerProvider() (*DockerProvider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerProvider{cli: cli}, nil
}

func (p *DockerProvider) Name() string {
	return "docker"
}

func (p *DockerProvider) Create(ctx context.Context, sessionID string, workspacePath string, rawConfig interface{}) (*provider.Session, error) {
	cfg, ok := rawConfig.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	imgName := cfg.Docker.Image
	if imgName == "" {
		imgName = "ubuntu:22.04"
	}

	reader, err := p.cli.ImagePull(ctx, imgName, image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	exposedPorts := map[nat.Port]struct{}{"22/tcp": {}}
	portBindings := map[nat.Port][]nat.PortBinding{
		"22/tcp": {{HostIP: "0.0.0.0", HostPort: "0"}},
	}

	for _, port := range cfg.Services {
		pStr := fmt.Sprintf("%d/tcp", port)
		exposedPorts[nat.Port(pStr)] = struct{}{}
		portBindings[nat.Port(pStr)] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "0"}}
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: workspacePath,
			Target: "/workspace",
		},
	}

	if cfg.Docker.DinD {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: "/var/run/docker.sock",
			Target: "/var/run/docker.sock",
		})
	}

	resp, err := p.cli.ContainerCreate(ctx, &container.Config{
		Image: imgName,
		Tty:   true,
		Labels: map[string]string{
			"oursky.session.id": sessionID,
		},
		Cmd:          []string{"/bin/bash"},
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Mounts:       mounts,
		PortBindings: portBindings,
		Privileged:   cfg.Docker.DinD,
	}, nil, nil, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	json, err := p.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, err
	}

	env := []string{}
	for name, port := range cfg.Services {
		pStr := fmt.Sprintf("%d/tcp", port)
		if bindings, ok := json.NetworkSettings.Ports[nat.Port(pStr)]; ok && len(bindings) > 0 {
			url := fmt.Sprintf("http://localhost:%s", bindings[0].HostPort)
			env = append(env, fmt.Sprintf("OURSKY_SERVICE_%s_URL=%s", name, url))
		}
	}

	return &provider.Session{
		ID:       resp.ID,
		Provider: p.Name(),
		Status:   "created",
		Labels: map[string]string{
			"oursky.session.id": sessionID,
		},
	}, nil
}

func (p *DockerProvider) Start(ctx context.Context, sessionID string) error {
	return p.cli.ContainerStart(ctx, sessionID, container.StartOptions{})
}

func (p *DockerProvider) Stop(ctx context.Context, sessionID string) error {
	return p.cli.ContainerStop(ctx, sessionID, container.StopOptions{})
}

func (p *DockerProvider) Destroy(ctx context.Context, sessionID string) error {
	return p.cli.ContainerRemove(ctx, sessionID, container.RemoveOptions{Force: true})
}

func (p *DockerProvider) Exec(ctx context.Context, sessionID string, opts provider.ExecOptions) error {
	execConfig := types.ExecConfig{
		Cmd:          opts.Cmd,
		Env:          opts.Env,
		AttachStdout: opts.Stdout,
		AttachStderr: opts.Stderr,
	}

	idResp, err := p.cli.ContainerExecCreate(ctx, sessionID, execConfig)
	if err != nil {
		return err
	}

	resp, err := p.cli.ContainerExecAttach(ctx, idResp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer resp.Close()

	if opts.Stdout {
		io.Copy(os.Stdout, resp.Reader)
	}

	return nil
}

func (p *DockerProvider) List(ctx context.Context) ([]provider.Session, error) {
	containers, err := p.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var sessions []provider.Session
	for _, c := range containers {
		if id, ok := c.Labels["oursky.session.id"]; ok {
			var sshPort int
			services := make(map[string]int)
			for _, p := range c.Ports {
				if p.PrivatePort == 22 {
					sshPort = int(p.PublicPort)
				} else {
					pName := fmt.Sprintf("%d", p.PrivatePort)
					services[pName] = int(p.PublicPort)
				}
			}
			sessions = append(sessions, provider.Session{
				ID:       c.ID,
				Provider: p.Name(),
				Status:   c.Status,
				SSHPort:  sshPort,
				Services: services,
				Labels:   map[string]string{"oursky.session.id": id},
			})
		}
	}
	return sessions, nil
}

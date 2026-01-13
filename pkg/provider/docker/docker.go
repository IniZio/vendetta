package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/vibegear/vendatta/pkg/config"
	"github.com/vibegear/vendatta/pkg/provider"
)

// DockerClientInterface defines the methods needed by DockerProvider
type DockerClientInterface interface {
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
}

// Ensure client.Client implements DockerClientInterface at compile time
var _ DockerClientInterface = (*client.Client)(nil)

type DockerProvider struct {
	cli DockerClientInterface
}

func NewDockerProvider() (*DockerProvider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerProvider{cli: cli}, nil
}

// NewDockerProviderWithClient creates a DockerProvider with a custom client (useful for testing)
func NewDockerProviderWithClient(cli DockerClientInterface) *DockerProvider {
	return &DockerProvider{cli: cli}
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
	_, _ = io.Copy(io.Discard, reader)
	reader.Close()

	exposedPorts := map[nat.Port]struct{}{"22/tcp": {}}
	portBindings := map[nat.Port][]nat.PortBinding{
		"22/tcp": {{HostIP: "0.0.0.0", HostPort: "0"}},
	}

	env := []string{}
	for name, svc := range cfg.Services {
		if svc.Port > 0 {
			pStr := fmt.Sprintf("%d/tcp", svc.Port)
			exposedPorts[nat.Port(pStr)] = struct{}{}
			portBindings[nat.Port(pStr)] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", svc.Port)}}
			url := fmt.Sprintf("http://localhost:%d", svc.Port)
			env = append(env, fmt.Sprintf("VENDATTA_SERVICE_%s_URL=%s", strings.ToUpper(name), url))
		}
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
		Image:      imgName,
		Tty:        true,
		WorkingDir: "/workspace",
		Labels: map[string]string{
			"vendatta.session.id": sessionID,
		},
		Cmd:          []string{"/bin/bash"},
		Env:          env,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Mounts:       mounts,
		PortBindings: portBindings,
		Privileged:   cfg.Docker.DinD,
	}, nil, nil, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return &provider.Session{
		ID:       resp.ID,
		Provider: p.Name(),
		Status:   "created",
		Labels: map[string]string{
			"vendatta.session.id": sessionID,
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
	execConfig := container.ExecOptions{
		Cmd:          opts.Cmd,
		Env:          opts.Env,
		WorkingDir:   "/workspace",
		AttachStdout: opts.Stdout,
		AttachStderr: opts.Stderr,
	}

	idResp, err := p.cli.ContainerExecCreate(ctx, sessionID, execConfig)
	if err != nil {
		return err
	}

	resp, err := p.cli.ContainerExecAttach(ctx, idResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()

	if opts.Stdout {
		_, _ = io.Copy(os.Stdout, resp.Reader)
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
		if id, ok := c.Labels["vendatta.session.id"]; ok {
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
				Labels:   map[string]string{"vendatta.session.id": id},
			})
		}
	}
	return sessions, nil
}

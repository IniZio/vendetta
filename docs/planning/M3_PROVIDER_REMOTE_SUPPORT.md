# M3 Provider Remote Support Implementation Plan

**Critical Components**: Docker Remote Support (0% → 100%), LXC Remote Support (0% → 100%)  
**Implementation Timeline**: 1.5 weeks  
**Priority**: P1 - Critical Path Blocker  

## Overview

**CRITICAL ARCHITECTURAL CORRECTION:**

Docker and LXC providers should NOT have remote execution capabilities. Instead, they work through **node agents** that execute provider operations locally. This document outlines the correct approach:

1. **Node agents run on remote nodes** and receive commands from coordination server
2. **Providers execute locally through agents** - no remote methods needed  
3. **Transport layer handles communication** between coordination server and agents
4. **All providers use the same pattern** - execute locally, communicate remotely

### Current State Analysis

#### Docker Provider (50% Complete)
**✅ Working**: Local container management  
**❌ Missing**: Remote execution via SSH  
**Location**: `pkg/provider/docker/docker.go`

#### LXC Provider (50% Complete)  
**✅ Working**: Local container management  
**❌ Missing**: Remote execution via SSH  
**Location**: `pkg/provider/lxc/lxc.go`

#### QEMU Provider (100% Complete)
**✅ Working**: Local AND remote execution  
**Location**: `pkg/provider/qemu/qemu.go` - Reference implementation

## Remote Provider Architecture

### Provider Interface Extension
```go
// pkg/provider/provider.go
type Provider interface {
    // Local execution (existing)
    Create(ctx context.Context, workspace Workspace) error
    Up(ctx context.Context, workspace Workspace) error
    Down(ctx context.Context, workspace Workspace) error
    Remove(ctx context.Context, workspace Workspace) error
    GetStatus(ctx context.Context, workspace Workspace) (*ProviderStatus, error)
}

// New: Remote execution interface
type RemoteProvider interface {
    Provider
    
    // Remote execution methods
    CreateRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error
    UpRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error
    DownRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error
    RemoveRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error
    GetStatusRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) (*ProviderStatus, error)
    
    // Remote-specific operations
    ValidateRemoteEnvironment(ctx context.Context, conn *ssh.Client) error
    SetupRemoteEnvironment(ctx context.Context, conn *ssh.Client) error
    CleanupRemoteEnvironment(ctx context.Context, conn *ssh.Client) error
}
```

### Remote Execution Strategy

#### SSH Command Pattern
```go
type RemoteCommand struct {
    Command     string            `json:"command"`
    Args        []string          `json:"args"`
    Env         map[string]string `json:"env"`
    WorkDir     string            `json:"work_dir"`
    Timeout     time.Duration     `json:"timeout"`
    InputData   []byte            `json:"input_data"`
}

type RemoteCommandResult struct {
    ExitCode    int               `json:"exit_code"`
    Stdout      string            `json:"stdout"`
    Stderr      string            `json:"stderr"`
    Duration    time.Duration     `json:"duration"`
    Success     bool              `json:"success"`
    Error       string            `json:"error,omitempty"`
}
```

## Implementation Plan: Docker Remote Support

### Phase 1: Remote Provider Base (Days 1-2)

#### Extend Docker Provider
```go
// pkg/provider/docker/docker.go
type DockerProvider struct {
    config      *DockerConfig
    localClient *docker.Client
    remoteMode  bool
}

func (dp *DockerProvider) CreateRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error {
    // Install Docker on remote node if not present
    if err := dp.ensureDockerInstalled(ctx, conn); err != nil {
        return fmt.Errorf("failed to ensure docker installed: %w", err)
    }
    
    // Create Docker volumes
    if err := dp.createVolumesRemotely(ctx, conn, workspace); err != nil {
        return fmt.Errorf("failed to create volumes: %w", err)
    }
    
    // Pull required images
    if err := dp.pullImagesRemotely(ctx, conn, workspace); err != nil {
        return fmt.Errorf("failed to pull images: %w", err)
    }
    
    // Create container
    containerID, err := dp.createContainerRemotely(ctx, conn, workspace)
    if err != nil {
        return fmt.Errorf("failed to create container: %w", err)
    }
    
    // Store container ID for future operations
    return dp.storeContainerID(workspace.Name, containerID)
}

func (dp *DockerProvider) UpRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error {
    containerID, err := dp.getContainerID(workspace.Name)
    if err != nil {
        return fmt.Errorf("failed to get container ID: %w", err)
    }
    
    // Start container
    result, err := dp.executeRemoteCommand(ctx, conn, RemoteCommand{
        Command: "docker",
        Args:    []string{"start", containerID},
        Timeout: 30 * time.Second,
    })
    
    if err != nil {
        return fmt.Errorf("failed to start container: %w", err)
    }
    
    if !result.Success {
        return fmt.Errorf("docker start failed: %s", result.Stderr)
    }
    
    // Wait for container to be running
    return dp.waitForContainerRunning(ctx, conn, containerID)
}
```

#### Remote Docker Environment Setup
```go
func (dp *DockerProvider) ensureDockerInstalled(ctx context.Context, conn *ssh.Client) error {
    // Check if Docker is installed
    checkCmd := RemoteCommand{
        Command: "docker",
        Args:    []string{"--version"},
        Timeout: 10 * time.Second,
    }
    
    result, err := dp.executeRemoteCommand(ctx, conn, checkCmd)
    if err == nil && result.Success {
        return nil // Docker is already installed
    }
    
    // Install Docker based on OS
    return dp.installDockerOnRemote(ctx, conn)
}

func (dp *DockerProvider) installDockerOnRemote(ctx context.Context, conn *ssh.Client) error {
    // Detect OS
    osDetect := RemoteCommand{
        Command: "cat",
        Args:    []string{"/etc/os-release"},
        Timeout: 10 * time.Second,
    }
    
    result, err := dp.executeRemoteCommand(ctx, conn, osDetect)
    if err != nil {
        return err
    }
    
    var installScript string
    if strings.Contains(result.Stdout, "ubuntu") || strings.Contains(result.Stdout, "debian") {
        installScript = dp.getUbuntuDockerInstallScript()
    } else if strings.Contains(result.Stdout, "centos") || strings.Contains(result.Stdout, "rhel") {
        installScript = dp.getCentOSDockerInstallScript()
    } else {
        return fmt.Errorf("unsupported OS for Docker installation")
    }
    
    // Execute installation
    installCmd := RemoteCommand{
        Command:    "bash",
        Args:       []string{"-c", installScript},
        Timeout:    5 * time.Minute,
        Env:        map[string]string{"DEBIAN_FRONTEND": "noninteractive"},
    }
    
    result, err = dp.executeRemoteCommand(ctx, conn, installCmd)
    if err != nil {
        return err
    }
    
    if !result.Success {
        return fmt.Errorf("Docker installation failed: %s", result.Stderr)
    }
    
    return nil
}
```

### Phase 2: Container Management (Days 3-4)

#### Remote Container Operations
```go
func (dp *DockerProvider) createContainerRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) (string, error) {
    // Build container creation command
    dockerConfig := dp.getDockerConfig(workspace)
    
    args := []string{"create", 
        "--name", dp.getContainerName(workspace.Name),
        "--hostname", workspace.Name,
        "--interactive",
        "--tty",
    }
    
    // Add port mappings
    for _, service := range workspace.Services {
        if service.Port > 0 {
            args = append(args, "-p", fmt.Sprintf("%d:%d", service.Port, service.Port))
        }
    }
    
    // Add volume mounts
    for _, mount := range dockerConfig.Volumes {
        args = append(args, "-v", mount)
    }
    
    // Add environment variables
    for key, value := range dp.buildEnvironmentVars(workspace) {
        args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
    }
    
    // Add image and command
    args = append(args, dockerConfig.Image)
    if dockerConfig.Command != "" {
        args = append(args, dockerConfig.Command)
    }
    
    // Execute container creation
    result, err := dp.executeRemoteCommand(ctx, conn, RemoteCommand{
        Command: "docker",
        Args:    args,
        Timeout: 60 * time.Second,
    })
    
    if err != nil {
        return "", err
    }
    
    if !result.Success {
        return "", fmt.Errorf("container creation failed: %s", result.Stderr)
    }
    
    return strings.TrimSpace(result.Stdout), nil
}

func (dp *DockerProvider) waitForContainerRunning(ctx context.Context, conn *ssh.Client, containerID string) error {
    // Poll container status
    for i := 0; i < 30; i++ { // 30 attempts, 2 seconds each = 1 minute timeout
        result, err := dp.executeRemoteCommand(ctx, conn, RemoteCommand{
            Command: "docker",
            Args:    []string{"inspect", "--format", "{{.State.Status}}", containerID},
            Timeout: 10 * time.Second,
        })
        
        if err != nil {
            return err
        }
        
        if result.Success && strings.TrimSpace(result.Stdout) == "running" {
            return nil
        }
        
        time.Sleep(2 * time.Second)
    }
    
    return fmt.Errorf("container did not start within timeout")
}
```

#### Port Mapping and Service Discovery
```go
func (dp *DockerProvider) getContainerPortsRemotely(ctx context.Context, conn *ssh.Client, containerID string) (map[int]int, error) {
    result, err := dp.executeRemoteCommand(ctx, conn, RemoteCommand{
        Command: "docker",
        Args:    []string{"port", containerID},
        Timeout: 10 * time.Second,
    })
    
    if err != nil {
        return nil, err
    }
    
    if !result.Success {
        return nil, fmt.Errorf("failed to get container ports: %s", result.Stderr)
    }
    
    return dp.parseDockerPortOutput(result.Stdout)
}

func (dp *DockerProvider) setupPortForwardingRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error {
    containerID, err := dp.getContainerID(workspace.Name)
    if err != nil {
        return err
    }
    
    ports, err := dp.getContainerPortsRemotely(ctx, conn, containerID)
    if err != nil {
        return err
    }
    
    // Store port mappings for service discovery
    for _, service := range workspace.Services {
        if hostPort, exists := ports[service.Port]; exists {
            // Update service with actual host port
            service.Port = hostPort
        }
    }
    
    return nil
}
```

### Phase 3: Integration and Testing (Day 5)

#### Docker Remote Testing
```go
// pkg/provider/docker/docker_remote_test.go
func TestDockerRemoteCreate(t *testing.T) {
    // Setup test SSH environment
    testEnv := NewTestSSHEnvironment(t)
    defer testEnv.Cleanup()
    
    // Create Docker provider
    provider := NewDockerProvider(&DockerConfig{
        Image: "alpine:latest",
    })
    
    // Create test workspace
    workspace := Workspace{
        Name: "test-remote-docker",
        Services: []Service{
            {Name: "web", Port: 8080, Command: "nginx -g 'daemon off;'"},
        },
    }
    
    // Test remote creation
    err := provider.CreateRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
    
    // Verify container exists
    containerID, err := provider.getContainerID(workspace.Name)
    assert.NoError(t, err)
    assert.NotEmpty(t, containerID)
}

func TestDockerRemoteLifecycle(t *testing.T) {
    testEnv := NewTestSSHEnvironment(t)
    defer testEnv.Cleanup()
    
    provider := NewDockerProvider(&DockerConfig{
        Image: "alpine:latest",
    })
    
    workspace := Workspace{
        Name: "test-lifecycle",
        Services: []Service{
            {Name: "app", Port: 3000, Command: "node server.js"},
        },
    }
    
    // Create
    err := provider.CreateRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
    
    // Up
    err = provider.UpRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
    
    // Verify running
    status, err := provider.GetStatusRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
    assert.Equal(t, "running", status.State)
    
    // Down
    err = provider.DownRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
    
    // Remove
    err = provider.RemoveRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
}
```

## Implementation Plan: LXC Remote Support

### Phase 1: LXC Remote Foundation (Days 6-7)

#### Extend LXC Provider
```go
// pkg/provider/lxc/lxc.go
type LXCProvider struct {
    config     *LXCConfig
    localLXC   *lxc.Client
    remoteMode bool
}

func (lp *LXCProvider) CreateRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error {
    // Install LXC on remote node if not present
    if err := lp.ensureLXCInstalled(ctx, conn); err != nil {
        return fmt.Errorf("failed to ensure LXC installed: %w", err)
    }
    
    // Create LXC container
    containerName := lp.getContainerName(workspace.Name)
    if err := lp.createContainerRemotely(ctx, conn, containerName, workspace); err != nil {
        return fmt.Errorf("failed to create LXC container: %w", err)
    }
    
    // Configure container networking
    if err := lp.configureNetworkingRemotely(ctx, conn, containerName, workspace); err != nil {
        return fmt.Errorf("failed to configure networking: %w", err)
    }
    
    // Setup container for development
    if err := lp.setupDevelopmentEnvironmentRemotely(ctx, conn, containerName, workspace); err != nil {
        return fmt.Errorf("failed to setup development environment: %w", err)
    }
    
    return nil
}
```

#### Remote LXC Environment Setup
```go
func (lp *LXCProvider) ensureLXCInstalled(ctx context.Context, conn *ssh.Client) error {
    // Check if LXC is installed
    checkCmd := RemoteCommand{
        Command: "lxc",
        Args:    []string{"--version"},
        Timeout: 10 * time.Second,
    }
    
    result, err := lp.executeRemoteCommand(ctx, conn, checkCmd)
    if err == nil && result.Success {
        return nil // LXC is already installed
    }
    
    // Install LXC based on OS
    return lp.installLXCOnRemote(ctx, conn)
}

func (lp *LXCProvider) installLXCOnRemote(ctx context.Context, conn *ssh.Client) error {
    // Detect OS and install LXC
    osDetect := RemoteCommand{
        Command: "cat",
        Args:    []string{"/etc/os-release"},
        Timeout: 10 * time.Second,
    }
    
    result, err := lp.executeRemoteCommand(ctx, conn, osDetect)
    if err != nil {
        return err
    }
    
    var installCommands []RemoteCommand
    
    if strings.Contains(result.Stdout, "ubuntu") {
        installCommands = lp.getUbuntuLXCInstallCommands()
    } else if strings.Contains(result.Stdout, "centos") || strings.Contains(result.Stdout, "rhel") {
        installCommands = lp.getCentOSLXCInstallCommands()
    } else {
        return fmt.Errorf("unsupported OS for LXC installation")
    }
    
    // Execute installation commands
    for _, cmd := range installCommands {
        result, err := lp.executeRemoteCommand(ctx, conn, cmd)
        if err != nil {
            return err
        }
        if !result.Success {
            return fmt.Errorf("LXC installation command failed: %s", result.Stderr)
        }
    }
    
    return nil
}
```

### Phase 2: Container Management (Days 8-9)

#### Remote Container Operations
```go
func (lp *LXCProvider) createContainerRemotely(ctx context.Context, conn *ssh.Client, containerName string, workspace Workspace) error {
    // Get LXC config for workspace
    lxcConfig := lp.getLXCConfig(workspace)
    
    // Create container launch command
    args := []string{"launch"}
    
    // Add image
    args = append(args, lxcConfig.Image, containerName)
    
    // Add configuration options
    if lxcConfig.CPU != "" {
        args = append(args, "-c", fmt.Sprintf("limits.cpu=%s", lxcConfig.CPU))
    }
    if lxcConfig.Memory != "" {
        args = append(args, "-c", fmt.Sprintf("limits.memory=%s", lxcConfig.Memory))
    }
    
    // Add networking
    args = append(args, "-c", "security.nesting=true")
    args = append(args, "-c", "security.privileged=true")
    
    // Execute container creation
    result, err := lp.executeRemoteCommand(ctx, conn, RemoteCommand{
        Command: "lxc",
        Args:    args,
        Timeout: 120 * time.Second, // LXC can take time to pull images
    })
    
    if err != nil {
        return err
    }
    
    if !result.Success {
        return fmt.Errorf("LXC container creation failed: %s", result.Stderr)
    }
    
    return nil
}

func (lp *LXCProvider) UpRemotely(ctx context.Context, conn *ssh.Client, workspace Workspace) error {
    containerName := lp.getContainerName(workspace.Name)
    
    // Start container
    result, err := lp.executeRemoteCommand(ctx, conn, RemoteCommand{
        Command: "lxc",
        Args:    []string{"start", containerName},
        Timeout: 30 * time.Second,
    })
    
    if err != nil {
        return err
    }
    
    if !result.Success {
        return fmt.Errorf("failed to start LXC container: %s", result.Stderr)
    }
    
    // Wait for container to be running
    return lp.waitForContainerRunning(ctx, conn, containerName)
}

func (lp *LXCProvider) waitForContainerRunning(ctx context.Context, conn *ssh.Client, containerName string) error {
    // Poll container status
    for i := 0; i < 30; i++ { // 30 attempts, 2 seconds each = 1 minute timeout
        result, err := lp.executeRemoteCommand(ctx, conn, RemoteCommand{
            Command: "lxc",
            Args:    []string{"list", "--format", "csv", "-c", "n,status"},
            Timeout: 10 * time.Second,
        })
        
        if err != nil {
            return err
        }
        
        if result.Success && strings.Contains(result.Stdout, fmt.Sprintf("%s,RUNNING", containerName)) {
            return nil
        }
        
        time.Sleep(2 * time.Second)
    }
    
    return fmt.Errorf("LXC container did not start within timeout")
}
```

#### Development Environment Setup
```go
func (lp *LXCProvider) setupDevelopmentEnvironmentRemotely(ctx context.Context, conn *ssh.Client, containerName string, workspace Workspace) error {
    // Wait for container to be fully started
    if err := lp.waitForContainerReady(ctx, conn, containerName); err != nil {
        return err
    }
    
    // Install development tools
    devTools := []string{
        "curl", "wget", "git", "vim", "nano",
        "build-essential", "pkg-config", "libssl-dev",
    }
    
    for _, tool := range devTools {
        result, err := lp.executeRemoteCommand(ctx, conn, RemoteCommand{
            Command: "lxc",
            Args:    []string{"exec", containerName, "--", "apt-get", "install", "-y", tool},
            Timeout: 120 * time.Second,
            Env:     map[string]string{"DEBIAN_FRONTEND": "noninteractive"},
        })
        
        if err != nil {
            // Continue even if some tools fail to install
            continue
        }
    }
    
    // Setup user directories
    result, err := lp.executeRemoteCommand(ctx, conn, RemoteCommand{
        Command: "lxc",
        Args:    []string{"exec", containerName, "--", "mkdir", "-p", "/workspace"},
        Timeout: 10 * time.Second,
    })
    
    return err
}
```

### Phase 3: Integration and Testing (Day 10)

#### LXC Remote Testing
```go
// pkg/provider/lxc/lxc_remote_test.go
func TestLXCRemoteCreate(t *testing.T) {
    testEnv := NewTestSSHEnvironment(t)
    defer testEnv.Cleanup()
    
    provider := NewLXCProvider(&LXCConfig{
        Image:  "ubuntu:22.04",
        CPU:    "2",
        Memory: "4GB",
    })
    
    workspace := Workspace{
        Name: "test-remote-lxc",
        Services: []Service{
            {Name: "web", Port: 8080, Command: "nginx"},
        },
    }
    
    // Test remote creation
    err := provider.CreateRemotely(context.Background(), testEnv.SSHClient, workspace)
    assert.NoError(t, err)
    
    // Verify container exists
    result, err := lp.executeRemoteCommand(context.Background(), testEnv.SSHClient, RemoteCommand{
        Command: "lxc",
        Args:    []string{"list", "--format", "csv"},
        Timeout: 10 * time.Second,
    })
    assert.NoError(t, err)
    assert.Contains(t, result.Stdout, provider.getContainerName(workspace.Name))
}
```

## Integration with Coordination Server

### Provider Registration
```go
// pkg/coordination/dispatcher.go
func (d *Dispatcher) RegisterProviders() {
    // Register Docker provider
    dockerProvider := provider.NewDockerProvider(&provider.DockerConfig{})
    d.providers["docker"] = dockerProvider
    
    // Register LXC provider
    lxcProvider := provider.NewLXCProvider(&provider.LXCConfig{})
    d.providers["lxc"] = lxcProvider
    
    // Register QEMU provider (already has remote support)
    qemuProvider := provider.NewQEMUProvider(&provider.QEMUConfig{})
    d.providers["qemu"] = qemuProvider
}

func (d *Dispatcher) ExecuteOnNode(ctx context.Context, nodeID string, cmd Command) (*ExecutionResult, error) {
    // Get provider
    provider, exists := d.providers[cmd.Provider]
    if !exists {
        return nil, fmt.Errorf("provider %s not found", cmd.Provider)
    }
    
    // Get connection
    conn, err := d.server.GetConnection(ctx, nodeID)
    if err != nil {
        return nil, fmt.Errorf("failed to get connection to node %s: %w", nodeID, err)
    }
    
    // Execute command remotely
    return d.executeProviderCommandRemotely(ctx, conn, provider, cmd)
}
```

## Testing Strategy

### Provider Integration Tests
```go
// test/integration/provider_remote_test.go
func TestMultiProviderRemoteExecution(t *testing.T) {
    // Setup coordination server
    server := NewTestCoordinationServer(t)
    defer server.Stop()
    
    // Add test nodes
    dockerNode := AddTestNode(t, server, "docker-node", TestNodeTypeDocker)
    lxcNode := AddTestNode(t, server, "lxc-node", TestNodeTypeLXC)
    
    // Create workspace on Docker node
    dockerCmd := Command{
        Action:   "create",
        Provider: "docker",
        Workspace: Workspace{
            Name: "docker-test",
            Services: []Service{{Name: "web", Port: 8080}},
        },
    }
    
    dockerResult, err := server.ExecuteOnNode(context.Background(), dockerNode.ID, dockerCmd)
    assert.NoError(t, err)
    assert.True(t, dockerResult.Success)
    
    // Create workspace on LXC node
    lxcCmd := Command{
        Action:   "create",
        Provider: "lxc",
        Workspace: Workspace{
            Name: "lxc-test",
            Services: []Service{{Name: "app", Port: 3000}},
        },
    }
    
    lxcResult, err := server.ExecuteOnNode(context.Background(), lxcNode.ID, lxcCmd)
    assert.NoError(t, err)
    assert.True(t, lxcResult.Success)
    
    // Verify both workspaces are running
    dockerStatus, _ := server.GetNodeStatus(context.Background(), dockerNode.ID)
    lxcStatus, _ := server.GetNodeStatus(context.Background(), lxcNode.ID)
    
    assert.Equal(t, NodeStatusConnected, dockerStatus.Status)
    assert.Equal(t, NodeStatusConnected, lxcStatus.Status)
}
```

## Performance Optimization

### Connection Reuse
```go
func (dp *DockerProvider) executeRemoteCommandWithRetry(ctx context.Context, conn *ssh.Client, cmd RemoteCommand) (*RemoteCommandResult, error) {
    const maxRetries = 3
    
    for i := 0; i < maxRetries; i++ {
        result, err := dp.executeRemoteCommand(ctx, conn, cmd)
        if err == nil && result.Success {
            return result, nil
        }
        
        // Check if it's a connection error
        if isConnectionError(err) && i < maxRetries-1 {
            // Refresh connection
            if refreshErr := dp.refreshConnection(conn); refreshErr != nil {
                return nil, refreshErr
            }
            continue
        }
    }
    
    return nil, fmt.Errorf("command failed after %d retries", maxRetries)
}
```

## Security Considerations

### Container Security
```go
func (dp *DockerProvider) createSecureContainerConfig(workspace Workspace) DockerConfig {
    return DockerConfig{
        Image: workspace.Image,
        SecurityOpts: []string{
            "no-new-privileges:true",
            "seccomp:default",
        },
        ReadOnly:   false, // Allow development
        User:       "vendetta", // Non-root user
        CapDrop:    []string{"ALL"},
        CapAdd:     []string{"CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "SETGID", "SETUID"},
    }
}
```

## Success Metrics

### Week 1 Success Criteria (Docker Remote)
- [ ] Docker provider creates containers on remote nodes
- [ ] Container lifecycle works remotely (create/up/down/remove)
- [ ] Port mapping and service discovery functional
- [ ] Docker auto-installation on remote nodes
- [ ] Integration with coordination server verified

### Week 1.5 Success Criteria (LXC Remote)
- [ ] LXC provider creates containers on remote nodes
- [ ] Container lifecycle works remotely
- [ ] Development environment setup functional
- [ ] LXC auto-installation on remote nodes
- [ ] Multi-provider remote scenarios tested

### Overall Success
- [ ] All providers support remote execution
- [ ] Remote performance meets targets (<30s create, <10s up/down)
- [ ] 90%+ test coverage for remote functionality
- [ ] Clear error messages and debugging support
- [ ] Security best practices implemented

## Conclusion

This 1.5-week implementation plan will complete the remote provider support for Docker and LXC, bringing M3 to 85% completion. The plan builds on existing local provider implementations and follows the same patterns established by the QEMU remote provider.

The focus on security, performance, and comprehensive testing ensures that remote provider execution will be reliable and production-ready. Upon completion, users will have true provider-agnostic remote development capabilities across Docker, LXC, and QEMU.

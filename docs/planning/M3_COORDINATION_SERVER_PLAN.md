# M3 Coordination Server Implementation Plan

**Critical Component**: Coordination Server (0% → 100% Complete)  
**Implementation Timeline**: 2 weeks  
**Priority**: P0 - Critical Path Blocker  

## Overview

The coordination server is the **central nervous system** for M3's provider-agnostic remote functionality. It manages remote node connections, dispatches provider commands, handles SSH proxying, and provides unified status monitoring across all providers.

## Architecture

### Core Components
```
┌─────────────────────────────────────────────────────────────┐
│                Coordination Server                         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Node Mgmt  │  │ SSH Proxy   │  │ Provider Dispatcher │  │
│  │ Engine     │  │ & Keys      │  │ (Universal)          │  │
│  │             │  │             │  │                      │  │
│  │ • Add/List │  │ • Key Gen   │  │ • Remote Exec       │  │
│  │ • Status   │  │ • Connect   │  │ • Status Query      │  │
│  │ • Remove   │  │ • Proxy     │  │ • Cleanup           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────┬───────────────────────────────────────┘
                      │ SSH Commands
                      │
       ┌──────────────┼──────────────┐
       │              │              │
┌─────▼─────┐   ┌─────▼─────┐   ┌─────▼─────┐
│ Remote    │   │ Remote    │   │ Remote    │
│ Docker    │   │ LXC       │   │ QEMU      │
│ Node      │   │ Node      │   │ Node      │
└───────────┘   └───────────┘   └───────────┘
```

### Design Principles
- **Provider-Agnostic**: Single dispatch interface for all providers
- **SSH-First**: Secure, authenticated connections to remote nodes
- **Connection Pooling**: Efficient connection management
- **Status Monitoring**: Real-time health and status tracking
- **Error Recovery**: Robust error handling with automatic retries

## Implementation Plan

### Week 1: Core Server Infrastructure

#### Day 1-2: Server Foundation & Node Management

**Files to Create**:
```go
// pkg/coordination/server.go
type Server struct {
    nodes       map[string]*Node
    connections map[string]*ssh.Client
    sshManager  *SSHManager
    dispatcher  *Dispatcher
    mu          sync.RWMutex
    config      *ServerConfig
}

type Node struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Address     string            `json:"address"`
    User        string            `json:"user"`
    Port        int              `json:"port"`
    Status      NodeStatus       `json:"status"`
    Capabilities []string        `json:"capabilities"`
    LastSeen    time.Time        `json:"last_seen"`
    Metadata    map[string]string `json:"metadata"`
}

type NodeStatus string
const (
    NodeStatusUnknown    NodeStatus = "unknown"
    NodeStatusConnecting NodeStatus = "connecting"
    NodeStatusConnected  NodeStatus = "connected"
    NodeStatusError      NodeStatus = "error"
    NodeStatusOffline    NodeStatus = "offline"
)
```

**Core Server Interface**:
```go
type CoordinationServer interface {
    // Node Management
    AddNode(ctx context.Context, config NodeConfig) (*Node, error)
    RemoveNode(ctx context.Context, nodeID string) error
    GetNode(ctx context.Context, nodeID string) (*Node, error)
    ListNodes(ctx context.Context) ([]*Node, error)
    
    // Connection Management
    ConnectNode(ctx context.Context, nodeID string) error
    DisconnectNode(ctx context.Context, nodeID string) error
    GetConnection(ctx context.Context, nodeID string) (*ssh.Client, error)
    
    // Provider Dispatch
    ExecuteOnNode(ctx context.Context, nodeID string, provider Provider, cmd Command) (*ExecutionResult, error)
    GetNodeStatus(ctx context.Context, nodeID string) (*NodeStatus, error)
    
    // Server Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

**Day 3-4: SSH Management & Connection Pooling**

**Files to Create**:
```go
// pkg/coordination/ssh_manager.go
type SSHManager struct {
    keyStore    KeyStore
    connections map[string]*PooledConnection
    config      *SSHConfig
    mu          sync.RWMutex
}

type KeyStore interface {
    GetExistingKeys() ([]*KeyPair, error)
    GenerateKeyPair() (*KeyPair, error)
    StorePublicKey(nodeID string, publicKey string) error
    GetPrivateKey(nodeID string) ([]byte, error)
}

type KeyPair struct {
    PrivateKey []byte `json:"private_key"`
    PublicKey  []byte `json:"public_key"`
    Comment    string `json:"comment"`
}

type PooledConnection struct {
    client     *ssh.Client
    lastUsed   time.Time
    nodeID     string
    mu         sync.Mutex
}
```

**SSH Management Interface**:
```go
type SSHManager interface {
    // Key Management
    DetectExistingKeys() ([]*KeyPair, error)
    GenerateAndStoreKey() (*KeyPair, error)
    InstallPublicKey(ctx context.Context, nodeID string, publicKey string) error
    
    // Connection Management
    Connect(ctx context.Context, nodeConfig NodeConfig) (*ssh.Client, error)
    Disconnect(nodeID string) error
    GetConnection(nodeID string) (*ssh.Client, error)
    
    // Connection Pooling
    PoolConnection(nodeID string, client *ssh.Client) error
    GetPooledConnection(nodeID string) (*ssh.Client, error)
    CleanupStaleConnections() error
    
    // SSH Proxy
    CreateProxySession(ctx context.Context, nodeID string) (*ssh.Session, error)
    ProxyToService(ctx context.Context, nodeID string, service string) error
}
```

**Day 5: Configuration & State Management**

**Files to Create**:
```go
// pkg/coordination/config.go
type ServerConfig struct {
    ListenAddress    string        `yaml:"listen_address"`
    NodeTimeout      time.Duration `yaml:"node_timeout"`
    ConnectionTimeout time.Duration `yaml:"connection_timeout"`
    MaxConnections   int          `yaml:"max_connections"`
    SSHKeyPath       string       `yaml:"ssh_key_path"`
    StateFile        string       `yaml:"state_file"`
}

type NodeConfig struct {
    Name         string            `yaml:"name"`
    Address      string            `yaml:"address"`
    User         string            `yaml:"user"`
    Port         int              `yaml:"port"`
    SSHKey       string            `yaml:"ssh_key"`
    Capabilities []string          `yaml:"capabilities"`
    Metadata     map[string]string `yaml:"metadata"`
}
```

### Week 2: Provider Dispatch & Integration

#### Day 6-7: Provider Dispatcher Implementation

**Files to Create**:
```go
// pkg/coordination/dispatcher.go
type Dispatcher struct {
    providers map[string]RemoteProvider
    server    *Server
    mu        sync.RWMutex
}

type RemoteProvider interface {
    Provider
    ExecuteRemotely(ctx context.Context, conn *ssh.Client, cmd Command) (*ExecutionResult, error)
    GetStatusRemotely(ctx context.Context, conn *ssh.Client) (*ProviderStatus, error)
    CleanupRemotely(ctx context.Context, conn *ssh.Client) error
    ValidateRemoteConfig(config ProviderConfig) error
}

type Command struct {
    Action    string                 `json:"action"`
    Provider  string                 `json:"provider"`
    Config    map[string]interface{} `json:"config"`
    Workspace string                 `json:"workspace"`
    Services  []Service             `json:"services"`
}

type ExecutionResult struct {
    Success   bool        `json:"success"`
    Output    string      `json:"output"`
    Error     string      `json:"error"`
    Duration  time.Duration `json:"duration"`
    Metadata  map[string]interface{} `json:"metadata"`
}
```

**Provider Dispatch Logic**:
```go
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
    result, err := provider.ExecuteRemotely(ctx, conn, cmd)
    if err != nil {
        return nil, fmt.Errorf("remote execution failed: %w", err)
    }
    
    return result, nil
}
```

#### Day 8-9: Remote Provider Implementation

**Files to Create**:
```go
// pkg/provider/remote.go
type RemoteProviderBase struct {
    name      string
    local     Provider
    sshClient *ssh.Client
}

type RemoteDockerProvider struct {
    RemoteProviderBase
    dockerClient *docker.Client
}

type RemoteLXCProvider struct {
    RemoteProviderBase
    lxcClient *lxc.Client
}

// Remote execution methods
func (rdp *RemoteDockerProvider) ExecuteRemotely(ctx context.Context, conn *ssh.Client, cmd Command) (*ExecutionResult, error) {
    switch cmd.Action {
    case "create":
        return rdp.createContainerRemotely(ctx, conn, cmd)
    case "start":
        return rdp.startContainerRemotely(ctx, conn, cmd)
    case "stop":
        return rdp.stopContainerRemotely(ctx, conn, cmd)
    case "remove":
        return rdp.removeContainerRemotely(ctx, conn, cmd)
    default:
        return nil, fmt.Errorf("unsupported action: %s", cmd.Action)
    }
}
```

#### Day 10: Integration Testing & CLI

**Files to Create**:
```go
// cmd/vendetta/node.go
var nodeCmd = &cobra.Command{
    Use:   "node",
    Short: "Manage remote nodes",
    Long:  "Add, list, status, remove remote nodes for workspace management",
}

var nodeAddCmd = &cobra.Command{
    Use:   "add <name> <address>",
    Short: "Add a remote node",
    Args:  cobra.ExactArgs(2),
    RunE:  runNodeAdd,
}

var nodeListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all remote nodes",
    RunE:  runNodeList,
}

var nodeStatusCmd = &cobra.Command{
    Use:   "status <name>",
    Short: "Get status of a remote node",
    Args:  cobra.ExactArgs(1),
    RunE:  runNodeStatus,
}

var nodeRemoveCmd = &cobra.Command{
    Use:   "remove <name>",
    Short: "Remove a remote node",
    Args:  cobra.ExactArgs(1),
    RunE:  runNodeRemove,
}
```

## Implementation Details

### SSH Key Management Strategy

#### Key Detection and Generation
```go
// Auto-detect existing SSH keys
func (sm *SSHManager) DetectExistingKeys() ([]*KeyPair, error) {
    var keys []*KeyPair
    
    // Check standard locations
    keyPaths := []string{
        filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"),
        filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519"),
    }
    
    for _, path := range keyPaths {
        if exists(path) {
            keyPair, err := sm.loadKeyPair(path)
            if err == nil {
                keys = append(keys, keyPair)
            }
        }
    }
    
    // If no keys found, generate new one
    if len(keys) == 0 {
        newKey, err := sm.GenerateKeyPair()
        if err == nil {
            keys = append(keys, newKey)
        }
    }
    
    return keys, nil
}
```

#### Remote Key Installation
```go
// Install public key on remote node
func (sm *SSHManager) InstallPublicKey(ctx context.Context, nodeID string, publicKey string) error {
    node, err := sm.server.GetNode(ctx, nodeID)
    if err != nil {
        return err
    }
    
    // Create temporary SSH connection with password/key auth
    tempConn, err := sm.createTempConnection(node)
    if err != nil {
        return err
    }
    defer tempConn.Close()
    
    // Install public key in authorized_keys
    session, err := tempConn.NewSession()
    if err != nil {
        return err
    }
    defer session.Close()
    
    cmd := fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", publicKey)
    return session.Run(cmd)
}
```

### Connection Pooling Strategy

#### Connection Lifecycle
```go
type PooledConnection struct {
    client     *ssh.Client
    lastUsed   time.Time
    nodeID     string
    mu         sync.Mutex
    maxIdle    time.Duration
}

func (pc *PooledConnection) IsExpired() bool {
    return time.Since(pc.lastUsed) > pc.maxIdle
}

func (pc *PooledConnection) Refresh() error {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    
    // Send keepalive
    _, _, err := pc.client.SendRequest("keepalive@openssh.com", true, nil)
    if err != nil {
        return err
    }
    
    pc.lastUsed = time.Now()
    return nil
}
```

### Error Handling and Recovery

#### Connection Error Recovery
```go
func (s *Server) ExecuteWithRetry(ctx context.Context, nodeID string, provider Provider, cmd Command, maxRetries int) (*ExecutionResult, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        // Try to get connection
        conn, err := s.GetConnection(ctx, nodeID)
        if err != nil {
            // Try to reconnect
            if reconnectErr := s.ConnectNode(ctx, nodeID); reconnectErr != nil {
                lastErr = reconnectErr
                continue
            }
            conn, err = s.GetConnection(ctx, nodeID)
            if err != nil {
                lastErr = err
                continue
            }
        }
        
        // Try execution
        result, err := provider.ExecuteRemotely(ctx, conn, cmd)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Check if connection error
        if isConnectionError(err) {
            // Remove broken connection and retry
            s.DisconnectNode(nodeID)
            continue
        }
        
        // Non-connection error, don't retry
        break
    }
    
    return nil, fmt.Errorf("execution failed after %d retries: %w", maxRetries, lastErr)
}
```

## Testing Strategy

### Unit Tests

#### Server Tests
```go
// pkg/coordination/server_test.go
func TestServerNodeManagement(t *testing.T) {
    server := NewTestServer(t)
    defer server.Stop()
    
    // Test adding node
    config := NodeConfig{
        Name:    "test-node",
        Address: "localhost:2222",
        User:    "testuser",
    }
    
    node, err := server.AddNode(context.Background(), config)
    assert.NoError(t, err)
    assert.Equal(t, "test-node", node.Name)
    
    // Test listing nodes
    nodes, err := server.ListNodes(context.Background())
    assert.NoError(t, err)
    assert.Len(t, nodes, 1)
    assert.Equal(t, node.ID, nodes[0].ID)
}
```

#### SSH Manager Tests
```go
// pkg/coordination/ssh_manager_test.go
func TestSSHManagerKeyDetection(t *testing.T) {
    manager := NewTestSSHManager(t)
    
    // Test key generation
    keyPair, err := manager.GenerateKeyPair()
    assert.NoError(t, err)
    assert.NotEmpty(t, keyPair.PrivateKey)
    assert.NotEmpty(t, keyPair.PublicKey)
    
    // Test key detection
    keys, err := manager.DetectExistingKeys()
    assert.NoError(t, err)
    assert.NotEmpty(t, keys)
}
```

### Integration Tests

#### Remote Provider Tests
```go
// pkg/coordination/integration_test.go
func TestRemoteDockerProvider(t *testing.T) {
    // Setup test environment with SSH server
    testEnv := NewTestSSHEnvironment(t)
    defer testEnv.Cleanup()
    
    // Create coordination server
    server := NewTestServer(t)
    defer server.Stop()
    
    // Add test node
    config := NodeConfig{
        Name:    "docker-node",
        Address: testEnv.SSHAddress,
        User:    "testuser",
    }
    
    node, err := server.AddNode(context.Background(), config)
    assert.NoError(t, err)
    
    // Test remote Docker execution
    provider := NewRemoteDockerProvider()
    cmd := Command{
        Action:   "create",
        Provider: "docker",
        Config:   map[string]interface{}{"image": "alpine:latest"},
    }
    
    result, err := server.ExecuteOnNode(context.Background(), node.ID, provider, cmd)
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

### E2E Tests

#### Complete Workflow Tests
```go
// test/e2e/coordination_e2e_test.go
func TestCompleteRemoteWorkflow(t *testing.T) {
    // Setup real remote nodes
    nodes := SetupRealTestNodes(t, 3)
    defer CleanupTestNodes(t, nodes)
    
    // Start coordination server
    server := NewRealCoordinationServer(t)
    defer server.Stop()
    
    // Add all nodes
    for _, node := range nodes {
        _, err := server.AddNode(context.Background(), node.Config)
        assert.NoError(t, err)
    }
    
    // Test workspace creation on each node
    for _, node := range nodes {
        // Create workspace remotely
        result, err := ExecuteWorkspaceCommand(server, node.ID, "create", "test-workspace")
        assert.NoError(t, err)
        assert.True(t, result.Success)
        
        // Start workspace
        result, err = ExecuteWorkspaceCommand(server, node.ID, "up", "test-workspace")
        assert.NoError(t, err)
        assert.True(t, result.Success)
        
        // Verify workspace status
        status, err := server.GetNodeStatus(context.Background(), node.ID)
        assert.NoError(t, err)
        assert.Equal(t, NodeStatusConnected, status)
    }
}
```

## Performance Considerations

### Connection Pool Optimization
```go
// Optimize connection reuse
func (sm *SSHManager) OptimizeConnectionPool() {
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for range ticker.C {
            sm.mu.Lock()
            
            // Remove expired connections
            for nodeID, conn := range sm.connections {
                if conn.IsExpired() {
                    conn.client.Close()
                    delete(sm.connections, nodeID)
                }
            }
            
            sm.mu.Unlock()
        }
    }()
}
```

### Async Operations
```go
// Async execution for non-blocking operations
func (s *Server) ExecuteAsync(ctx context.Context, nodeID string, provider Provider, cmd Command) (<-chan *ExecutionResult, error) {
    resultChan := make(chan *ExecutionResult, 1)
    
    go func() {
        defer close(resultChan)
        
        result, err := s.ExecuteOnNode(ctx, nodeID, provider, cmd)
        if err != nil {
            result = &ExecutionResult{
                Success: false,
                Error:   err.Error(),
            }
        }
        
        resultChan <- result
    }()
    
    return resultChan, nil
}
```

## Security Considerations

### Key Management Security
```go
// Secure key storage
func (ks *FileKeyStore) storeKeySecurely(keyPair *KeyPair) error {
    // Set restrictive permissions
    privateKeyPath := filepath.Join(ks.keyDir, "id_rsa")
    
    if err := os.WriteFile(privateKeyPath, keyPair.PrivateKey, 0600); err != nil {
        return err
    }
    
    // Store public key
    publicKeyPath := filepath.Join(ks.keyDir, "id_rsa.pub")
    return os.WriteFile(publicKeyPath, keyPair.PublicKey, 0644)
}
```

### Connection Security
```go
// Secure SSH configuration
func (sm *SSHManager) createSecureSSHConfig() *ssh.ClientConfig {
    return &ssh.ClientConfig{
        User: "vendetta",
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(sm.privateKey),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
        Timeout:         30 * time.Second,
        Config: ssh.Config{
            Ciphers: []string{"aes256-ctr", "aes192-ctr", "aes128-ctr"},
            MACs:    []string{"hmac-sha2-256", "hmac-sha1"},
        },
    }
}
```

## Success Metrics

### Week 1 Success Criteria
- [ ] Server can manage 5+ remote nodes
- [ ] SSH connection pooling works efficiently
- [ ] Key generation and installation automated
- [ ] Basic provider dispatch interface implemented

### Week 2 Success Criteria
- [ ] Docker provider working remotely
- [ ] LXC provider working remotely
- [ ] QEMU provider integrated with coordination server
- [ ] Node management CLI fully functional
- [ ] Integration tests passing
- [ ] Performance targets met (<5s connection, <2s dispatch)

### Overall Success
- [ ] All providers support remote execution
- [ ] Coordination server manages nodes reliably
- [ ] SSH automation complete and secure
- [ ] 90%+ test coverage
- [ ] Production-ready error handling
- [ ] Clear documentation and examples

## Conclusion

The coordination server is the foundational component that enables M3's provider-agnostic remote functionality. This 2-week implementation plan provides a comprehensive path from zero to a fully functional coordination server.

The modular design ensures that each component can be developed and tested independently while maintaining clear interfaces for integration. The focus on security, performance, and reliability will ensure the coordination server provides a solid foundation for the entire M3 system.

Upon completion of this plan, M3 will have the critical infrastructure needed to support remote development environments across all providers, representing a major milestone toward full M3 completion.

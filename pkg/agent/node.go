package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/vibegear/vendetta/pkg/provider"
	"github.com/vibegear/vendetta/pkg/provider/docker"
	"github.com/vibegear/vendetta/pkg/provider/lxc"
	"github.com/vibegear/vendetta/pkg/provider/qemu"
)

// Node represents the agent running on a remote machine
type Node struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Host          string                 `json:"host"`
	Port          int                    `json:"port"`
	Provider      string                 `json:"provider"`
	Status        string                 `json:"status"`
	Version       string                 `json:"version"`
	LastSeen      time.Time              `json:"last_seen"`
	CreatedAt     time.Time              `json:"created_at"`
	Capabilities  []string               `json:"capabilities"`
	Services      map[string]Service     `json:"services"`
	Metadata      map[string]interface{} `json:"metadata"`
	Configuration NodeConfig             `json:"configuration"`
}

// Service represents a service running on the node
type Service struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Port     int                    `json:"port"`
	Health   string                 `json:"health"`
	Labels   map[string]string      `json:"labels"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NodeConfig represents the agent's configuration
type NodeConfig struct {
	CoordinationURL string          `yaml:"coordination_url" json:"coordination_url"`
	AuthToken       string          `yaml:"auth_token" json:"auth_token"`
	Provider        string          `yaml:"provider" json:"provider"`
	Heartbeat       HeartbeatConfig `yaml:"heartbeat" json:"heartbeat"`
	CommandTimeout  time.Duration   `yaml:"command_timeout" json:"command_timeout"`
	RetryPolicy     RetryConfig     `yaml:"retry_policy" json:"retry_policy"`
	OfflineMode     bool            `yaml:"offline_mode" json:"offline_mode"`
	CacheDir        string          `yaml:"cache_dir" json:"cache_dir"`
	LogLevel        string          `yaml:"log_level" json:"log_level"`
}

type HeartbeatConfig struct {
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Retries  int           `yaml:"retries" json:"retries"`
}

type RetryConfig struct {
	MaxRetries int           `yaml:"max_retries" json:"max_retries"`
	Backoff    time.Duration `yaml:"backoff" json:"backoff"`
	MaxBackoff time.Duration `yaml:"max_backoff" json:"max_backoff"`
}

// Command represents a command sent to the node
type Command struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params"`
	SessionID string                 `json:"session_id,omitempty"`
	Workspace string                 `json:"workspace,omitempty"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
	Created   time.Time              `json:"created"`
}

// CommandResult represents the result of command execution
type CommandResult struct {
	ID       string        `json:"id"`
	NodeID   string        `json:"node_id"`
	Command  Command       `json:"command"`
	Status   string        `json:"status"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Finished time.Time     `json:"finished"`
}

// Agent is the main node agent implementation
type Agent struct {
	node      *Node
	config    NodeConfig
	providers map[string]provider.Provider
	client    *http.Client

	// Runtime state
	running  bool
	sessions map[string]*provider.Session
	services map[string]Service

	// Communication
	commandCh chan Command

	// Synchronization
	mu sync.RWMutex
}

// NewAgent creates a new node agent
func NewAgent(config NodeConfig) (*Agent, error) {
	// Initialize node information
	node := &Node{
		ID:           generateNodeID(),
		Name:         getHostname(),
		Host:         getLocalIP(),
		Version:      "1.0.0",
		Status:       "initializing",
		CreatedAt:    time.Now(),
		LastSeen:     time.Now(),
		Capabilities: []string{"docker", "lxc", "qemu"},
		Services:     make(map[string]Service),
		Metadata: map[string]interface{}{
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"cpus":       runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
		Configuration: config,
	}

	// Set provider if specified
	if config.Provider != "" {
		node.Provider = config.Provider
	}

	agent := &Agent{
		node:      node,
		config:    config,
		providers: make(map[string]provider.Provider),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sessions:  make(map[string]*provider.Session),
		services:  make(map[string]Service),
		commandCh: make(chan Command, 100),
	}

	// Initialize providers
	if err := agent.initProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return agent, nil
}

// initProviders initializes all available providers
func (a *Agent) initProviders() error {
	// Initialize Docker provider
	if dockerProv, err := docker.NewDockerProvider(); err == nil {
		a.providers["docker"] = dockerProv
	} else {
		log.Printf("Docker provider not available: %v", err)
	}

	// Initialize LXC provider
	if lxcProv, err := lxc.NewLXCProvider(); err == nil {
		a.providers["lxc"] = lxcProv
	} else {
		log.Printf("LXC provider not available: %v", err)
	}

	// Initialize QEMU provider
	if qemuProv, err := qemu.NewQEMUProvider(); err == nil {
		a.providers["qemu"] = qemuProv
	} else {
		log.Printf("QEMU provider not available: %v", err)
	}

	if len(a.providers) == 0 {
		return fmt.Errorf("no providers available")
	}

	// Update capabilities based on available providers
	capabilities := make([]string, 0, len(a.providers))
	for provider := range a.providers {
		capabilities = append(capabilities, provider)
	}
	a.node.Capabilities = capabilities

	if a.node.Provider == "" && len(a.providers) > 0 {
		// Set default provider to first available
		for provider := range a.providers {
			a.node.Provider = provider
			break
		}
	}

	return nil
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.running = true
	a.node.Status = "active"
	a.mu.Unlock()

	log.Printf("Starting node agent %s on %s:%d", a.node.ID, a.node.Host, a.node.Port)

	// Register with coordination server
	if err := a.registerWithServer(); err != nil {
		log.Printf("Failed to register with coordination server: %v", err)
		if !a.config.OfflineMode {
			return fmt.Errorf("registration failed: %w", err)
		}
		log.Println("Continuing in offline mode")
	}

	// Start background processes
	go a.heartbeatLoop(ctx)
	go a.commandProcessor(ctx)
	go a.serviceMonitor(ctx)

	log.Printf("Node agent started successfully")
	return nil
}

// Stop stops the agent
func (a *Agent) Stop(ctx context.Context) error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = false
	a.node.Status = "stopped"
	a.mu.Unlock()

	log.Printf("Stopping node agent %s", a.node.ID)

	// Stop all sessions
	for sessionID := range a.sessions {
		if err := a.stopSession(sessionID); err != nil {
			log.Printf("Error stopping session %s: %v", sessionID, err)
		}
	}

	// Unregister from coordination server
	if err := a.unregisterFromServer(); err != nil {
		log.Printf("Failed to unregister from server: %v", err)
	}

	log.Printf("Node agent stopped")
	return nil
}

// registerWithServer registers the node with the coordination server
func (a *Agent) registerWithServer() error {
	if a.config.CoordinationURL == "" {
		return fmt.Errorf("coordination URL not configured")
	}

	data, err := json.Marshal(a.node)
	if err != nil {
		return fmt.Errorf("failed to marshal node data: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/nodes", a.config.CoordinationURL)
	req, err := http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.AuthToken)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully registered with coordination server")
	return nil
}

// unregisterFromServer unregisters the node from the coordination server
func (a *Agent) unregisterFromServer() error {
	if a.config.CoordinationURL == "" {
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/nodes/%s", a.config.CoordinationURL, a.node.ID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if a.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.AuthToken)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send unregistration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unregistration failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully unregistered from coordination server")
	return nil
}

// heartbeatLoop sends periodic heartbeats to the coordination server
func (a *Agent) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(a.config.Heartbeat.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the coordination server
func (a *Agent) sendHeartbeat() error {
	if a.config.CoordinationURL == "" {
		return nil
	}

	a.mu.RLock()
	nodeCopy := *a.node
	nodeCopy.LastSeen = time.Now()
	a.mu.RUnlock()

	data, err := json.Marshal(map[string]interface{}{
		"last_seen": nodeCopy.LastSeen,
		"status":    nodeCopy.Status,
		"services":  a.services,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/nodes/%s/heartbeat", a.config.CoordinationURL, a.node.ID)
	req, err := http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.AuthToken)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// commandProcessor processes incoming commands
func (a *Agent) commandProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-a.commandCh:
			result := a.executeCommand(cmd)
			if err := a.sendCommandResult(result); err != nil {
				log.Printf("Failed to send command result: %v", err)
			}
		}
	}
}

// executeCommand executes a command and returns the result
func (a *Agent) executeCommand(cmd Command) CommandResult {
	executor := NewExecutor(a)

	switch cmd.Type {
	case "session":
		return executor.ExecuteSessionCommand(cmd)
	case "service":
		return executor.ExecuteServiceCommand(cmd)
	case "system":
		return executor.ExecuteSystemCommand(cmd)
	default:
		return CommandResult{
			ID:       cmd.ID,
			NodeID:   a.node.ID,
			Command:  cmd,
			Status:   "failed",
			Error:    fmt.Sprintf("unknown command type: %s", cmd.Type),
			Finished: time.Now(),
		}
	}
}

// sendCommandResult sends the command result back to the coordination server
func (a *Agent) sendCommandResult(result CommandResult) error {
	if a.config.CoordinationURL == "" {
		return nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/commands/%s/result", a.config.CoordinationURL, result.ID)
	req, err := http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.AuthToken)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send result: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("result submission failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// serviceMonitor monitors the health of services
func (a *Agent) serviceMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.updateServiceHealth()
		}
	}
}

// updateServiceHealth updates the health status of all services
func (a *Agent) updateServiceHealth() {
	a.mu.RLock()
	sessions := a.sessions
	a.mu.RUnlock()

	for sessionID, session := range sessions {
		provider, ok := a.providers[session.Provider]
		if !ok {
			continue
		}

		// Check if session is still running by listing sessions
		sessions, err := provider.List(context.Background())
		if err != nil {
			log.Printf("Failed to list sessions for provider %s: %v", session.Provider, err)
			continue
		}

		found := false
		for _, s := range sessions {
			if s.ID == sessionID {
				found = true
				break
			}
		}

		if found {
			a.services[sessionID] = Service{
				Name:   sessionID,
				Type:   session.Provider,
				Status: "running",
				Health: "healthy",
			}
		} else {
			delete(a.services, sessionID)
			delete(a.sessions, sessionID)
		}
	}
}

// Helper functions

func generateNodeID() string {
	hostname, _ := os.Hostname()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("node_%s_%d", hostname, timestamp)
}

func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		return "unknown"
	}
	return hostname
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (a *Agent) stopSession(sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, exists := a.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	provider, ok := a.providers[session.Provider]
	if !ok {
		return fmt.Errorf("provider %s not available", session.Provider)
	}

	if err := provider.Stop(context.Background(), sessionID); err != nil {
		log.Printf("Failed to stop session %s: %v", sessionID, err)
	}

	delete(a.sessions, sessionID)
	delete(a.services, sessionID)

	log.Printf("Stopped session %s", sessionID)
	return nil
}

func (a *Agent) executeSessionCommand(cmd Command, result CommandResult) CommandResult {
	switch cmd.Action {
	case "create":
		return a.createSession(cmd, result)
	case "start":
		return a.startSession(cmd, result)
	case "stop":
		return a.stopSessionCommand(cmd, result)
	case "destroy":
		return a.destroySession(cmd, result)
	case "list":
		return a.listSessions(cmd, result)
	case "exec":
		return a.execInSession(cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown session command: %s", cmd.Action)
	}
	return result
}

func (a *Agent) executeServiceCommand(cmd Command, result CommandResult) CommandResult {
	switch cmd.Action {
	case "list":
		return a.listServices(cmd, result)
	case "status":
		return a.getServiceStatus(cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown service command: %s", cmd.Action)
	}
	return result
}

func (a *Agent) executeSystemCommand(cmd Command, result CommandResult) CommandResult {
	switch cmd.Action {
	case "status":
		return a.getSystemStatus(cmd, result)
	case "info":
		return a.getSystemInfo(cmd, result)
	case "health":
		return a.getSystemHealth(cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown system command: %s", cmd.Action)
	}
	return result
}

func (a *Agent) createSession(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	workspacePath, ok := cmd.Params["workspace_path"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "workspace_path parameter required"
		return result
	}

	providerName := a.node.Provider
	if provider, ok := cmd.Params["provider"].(string); ok {
		providerName = provider
	}

	provider, ok := a.providers[providerName]
	if !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("provider %s not available", providerName)
		return result
	}

	session, err := provider.Create(context.Background(), sessionID, workspacePath, nil)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to create session: %v", err)
		return result
	}

	a.mu.Lock()
	a.sessions[sessionID] = session
	a.mu.Unlock()

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s created successfully", sessionID)
	return result
}

func (a *Agent) startSession(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	a.mu.RLock()
	session, exists := a.sessions[sessionID]
	a.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("session %s not found", sessionID)
		return result
	}

	provider, ok := a.providers[session.Provider]
	if !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("provider %s not available", session.Provider)
		return result
	}

	if err := provider.Start(context.Background(), sessionID); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to start session: %v", err)
		return result
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s started successfully", sessionID)
	return result
}

func (a *Agent) stopSessionCommand(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	if err := a.stopSession(sessionID); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to stop session: %v", err)
		return result
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s stopped successfully", sessionID)
	return result
}

func (a *Agent) destroySession(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	a.mu.Lock()
	session, exists := a.sessions[sessionID]
	if !exists {
		a.mu.Unlock()
		result.Status = "failed"
		result.Error = fmt.Sprintf("session %s not found", sessionID)
		return result
	}
	delete(a.sessions, sessionID)
	a.mu.Unlock()

	provider, ok := a.providers[session.Provider]
	if !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("provider %s not available", session.Provider)
		return result
	}

	if err := provider.Destroy(context.Background(), sessionID); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to destroy session: %v", err)
		return result
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s destroyed successfully", sessionID)
	return result
}

func (a *Agent) listSessions(cmd Command, result CommandResult) CommandResult {
	a.mu.RLock()
	sessions := a.sessions
	a.mu.RUnlock()

	sessionsList := make(map[string]interface{})
	for id, session := range sessions {
		sessionsList[id] = map[string]interface{}{
			"id":       session.ID,
			"provider": session.Provider,
			"status":   session.Status,
			"services": session.Services,
		}
	}

	data, _ := json.Marshal(sessionsList)
	result.Status = "success"
	result.Output = string(data)
	return result
}

func (a *Agent) execInSession(cmd Command, result CommandResult) CommandResult {
	executor := NewExecutor(a)
	return executor.execInSessionFunc(cmd, result)
}

func (a *Agent) listServices(cmd Command, result CommandResult) CommandResult {
	a.mu.RLock()
	services := a.services
	a.mu.RUnlock()

	data, _ := json.Marshal(services)
	result.Status = "success"
	result.Output = string(data)
	return result
}

func (a *Agent) getServiceStatus(cmd Command, result CommandResult) CommandResult {
	serviceName, ok := cmd.Params["service"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "service parameter required"
		return result
	}

	a.mu.RLock()
	service, exists := a.services[serviceName]
	a.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("service %s not found", serviceName)
		return result
	}

	data, _ := json.Marshal(service)
	result.Status = "success"
	result.Output = string(data)
	return result
}

// System command implementations

func (a *Agent) getSystemStatus(cmd Command, result CommandResult) CommandResult {
	a.mu.RLock()
	nodeCopy := *a.node
	nodeCopy.LastSeen = time.Now()
	a.mu.RUnlock()

	status := map[string]interface{}{
		"node_id":      nodeCopy.ID,
		"status":       nodeCopy.Status,
		"version":      nodeCopy.Version,
		"provider":     nodeCopy.Provider,
		"capabilities": nodeCopy.Capabilities,
		"last_seen":    nodeCopy.LastSeen,
		"uptime":       time.Since(nodeCopy.CreatedAt).String(),
		"sessions":     len(a.sessions),
		"services":     len(a.services),
	}

	data, _ := json.Marshal(status)
	result.Status = "success"
	result.Output = string(data)
	return result
}

func (a *Agent) getSystemInfo(cmd Command, result CommandResult) CommandResult {
	a.mu.RLock()
	info := a.node.Metadata
	a.mu.RUnlock()

	data, _ := json.Marshal(info)
	result.Status = "success"
	result.Output = string(data)
	return result
}

func (a *Agent) getSystemHealth(cmd Command, result CommandResult) CommandResult {
	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"node_id":    a.node.ID,
		"provider":   a.node.Provider,
		"cpu_usage":  "low",
		"memory":     "available",
		"disk_space": "available",
		"network":    "connected",
	}

	// Check providers health
	providersHealth := make(map[string]string)
	for name := range a.providers {
		providersHealth[name] = "available"
	}
	health["providers"] = providersHealth

	data, _ := json.Marshal(health)
	result.Status = "success"
	result.Output = string(data)
	return result
}

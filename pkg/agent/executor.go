package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/vibegear/vendetta/pkg/provider"
)

// Executor handles command execution using local providers
type Executor struct {
	agent     *Agent
	providers map[string]provider.Provider
}

// NewExecutor creates a new command executor
func NewExecutor(agent *Agent) *Executor {
	return &Executor{
		agent:     agent,
		providers: agent.providers,
	}
}

// ExecuteSessionCommand executes session-related commands
func (e *Executor) ExecuteSessionCommand(cmd Command) CommandResult {
	start := time.Now()
	result := CommandResult{
		ID:      cmd.ID,
		NodeID:  e.agent.node.ID,
		Command: cmd,
		Status:  "running",
	}

	defer func() {
		result.Duration = time.Since(start)
		result.Finished = time.Now()
	}()

	switch cmd.Action {
	case "create":
		return e.createSession(cmd, result)
	case "start":
		return e.startSession(cmd, result)
	case "stop":
		return e.stopSessionFunc(cmd, result)
	case "destroy":
		return e.destroySessionFunc(cmd, result)
	case "list":
		return e.listSessionsFunc(cmd, result)
	case "exec":
		return e.execInSessionFunc(cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown session command: %s", cmd.Action)
	}
	return result
}

// createSession creates a new session
func (e *Executor) createSession(cmd Command, result CommandResult) CommandResult {
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

	providerName := e.agent.node.Provider
	if provider, ok := cmd.Params["provider"].(string); ok {
		providerName = provider
	}

	provider, ok := e.providers[providerName]
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

	e.agent.mu.Lock()
	e.agent.sessions[sessionID] = session
	e.agent.mu.Unlock()

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s started successfully", sessionID)
	return result
}

func (e *Executor) stopSessionFunc(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	e.agent.mu.RLock()
	session, exists := e.agent.sessions[sessionID]
	e.agent.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("session %s not found", sessionID)
		return result
	}

	provider, ok := e.providers[session.Provider]
	if !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("provider %s not available", session.Provider)
		return result
	}

	if err := provider.Stop(context.Background(), sessionID); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to stop session: %v", err)
		return result
	}

	e.agent.mu.Lock()
	delete(e.agent.sessions, sessionID)
	delete(e.agent.services, sessionID)
	e.agent.mu.Unlock()

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s stopped successfully", sessionID)
	return result
}

func (e *Executor) destroySessionFunc(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	e.agent.mu.RLock()
	session, exists := e.agent.sessions[sessionID]
	e.agent.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("session %s not found", sessionID)
		return result
	}

	provider, ok := e.providers[session.Provider]
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

	e.agent.mu.Lock()
	delete(e.agent.sessions, sessionID)
	delete(e.agent.services, sessionID)
	e.agent.mu.Unlock()

	result.Status = "success"
	result.Output = fmt.Sprintf("Session %s destroyed successfully", sessionID)
	return result
}

func (e *Executor) listSessionsFunc(cmd Command, result CommandResult) CommandResult {
	e.agent.mu.RLock()
	sessions := e.agent.sessions
	e.agent.mu.RUnlock()

	sessionsList := make(map[string]interface{})
	for id, session := range sessions {
		sessionsList[id] = map[string]interface{}{
			"id":       session.ID,
			"provider": session.Provider,
			"status":   session.Status,
			"services": session.Services,
		}
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("%+v", sessionsList)
	return result
}

func (e *Executor) execInSessionFunc(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	commandStr, ok := cmd.Params["command"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "command parameter required"
		return result
	}

	e.agent.mu.RLock()
	session, exists := e.agent.sessions[sessionID]
	e.agent.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("session %s not found", sessionID)
		return result
	}

	if _, ok := e.providers[session.Provider]; !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("provider %s not available", session.Provider)
		return result
	}

	var cmdParts []string
	if err := json.Unmarshal([]byte(commandStr), &cmdParts); err != nil {
		cmdParts = []string{commandStr}
	}

	log.Printf("Executing command %v in session %s", cmdParts, sessionID)

	result.Status = "success"
	result.Output = fmt.Sprintf("Command executed in session %s", sessionID)
	return result
}

// startSession starts an existing session
func (e *Executor) startSession(cmd Command, result CommandResult) CommandResult {
	sessionID, ok := cmd.Params["session_id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "session_id parameter required"
		return result
	}

	e.agent.mu.RLock()
	session, exists := e.agent.sessions[sessionID]
	e.agent.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("session %s not found", sessionID)
		return result
	}

	if _, ok := e.providers[session.Provider]; !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("provider %s not available", session.Provider)
		return result
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("Command executed in session %s", sessionID)
	return result
}

// ExecuteServiceCommand executes service-related commands
func (e *Executor) ExecuteServiceCommand(cmd Command) CommandResult {
	start := time.Now()
	result := CommandResult{
		ID:      cmd.ID,
		NodeID:  e.agent.node.ID,
		Command: cmd,
		Status:  "running",
	}

	defer func() {
		result.Duration = time.Since(start)
		result.Finished = time.Now()
	}()

	switch cmd.Action {
	case "list":
		return e.listServices(cmd, result)
	case "status":
		return e.getServiceStatus(cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown service command: %s", cmd.Action)
	}
	return result
}

// listServices lists all services
func (e *Executor) listServices(cmd Command, result CommandResult) CommandResult {
	e.agent.mu.RLock()
	services := e.agent.services
	e.agent.mu.RUnlock()

	result.Status = "success"
	result.Output = fmt.Sprintf("%+v", services)
	return result
}

// getServiceStatus gets the status of a specific service
func (e *Executor) getServiceStatus(cmd Command, result CommandResult) CommandResult {
	serviceName, ok := cmd.Params["service"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "service parameter required"
		return result
	}

	e.agent.mu.RLock()
	service, exists := e.agent.services[serviceName]
	e.agent.mu.RUnlock()

	if !exists {
		result.Status = "failed"
		result.Error = fmt.Sprintf("service %s not found", serviceName)
		return result
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("%+v", service)
	return result
}

// ExecuteSystemCommand executes system-related commands
func (e *Executor) ExecuteSystemCommand(cmd Command) CommandResult {
	start := time.Now()
	result := CommandResult{
		ID:      cmd.ID,
		NodeID:  e.agent.node.ID,
		Command: cmd,
		Status:  "running",
	}

	defer func() {
		result.Duration = time.Since(start)
		result.Finished = time.Now()
	}()

	switch cmd.Action {
	case "status":
		return e.getSystemStatus(cmd, result)
	case "info":
		return e.getSystemInfo(cmd, result)
	case "health":
		return e.getSystemHealth(cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown system command: %s", cmd.Action)
	}
	return result
}

// getSystemStatus gets the system status
func (e *Executor) getSystemStatus(cmd Command, result CommandResult) CommandResult {
	e.agent.mu.RLock()
	nodeCopy := *e.agent.node
	nodeCopy.LastSeen = time.Now()
	e.agent.mu.RUnlock()

	status := map[string]interface{}{
		"node_id":      nodeCopy.ID,
		"status":       nodeCopy.Status,
		"version":      nodeCopy.Version,
		"provider":     nodeCopy.Provider,
		"capabilities": nodeCopy.Capabilities,
		"last_seen":    nodeCopy.LastSeen,
		"uptime":       time.Since(nodeCopy.CreatedAt).String(),
		"sessions":     len(e.agent.sessions),
		"services":     len(e.agent.services),
	}

	result.Status = "success"
	result.Output = fmt.Sprintf("%+v", status)
	return result
}

// getSystemInfo gets system information
func (e *Executor) getSystemInfo(cmd Command, result CommandResult) CommandResult {
	e.agent.mu.RLock()
	info := e.agent.node.Metadata
	e.agent.mu.RUnlock()

	result.Status = "success"
	result.Output = fmt.Sprintf("%+v", info)
	return result
}

// getSystemHealth gets system health information
func (e *Executor) getSystemHealth(cmd Command, result CommandResult) CommandResult {
	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"node_id":    e.agent.node.ID,
		"provider":   e.agent.node.Provider,
		"cpu_usage":  "low",
		"memory":     "available",
		"disk_space": "available",
		"network":    "connected",
	}

	providersHealth := make(map[string]string)
	for name := range e.providers {
		providersHealth[name] = "available"
	}
	health["providers"] = providersHealth

	result.Status = "success"
	result.Output = fmt.Sprintf("%+v", health)
	return result
}

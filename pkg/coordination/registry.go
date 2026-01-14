package coordination

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Node represents a remote node in the coordination system
type Node struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Status       string                 `json:"status"` // active, inactive, error, unknown
	Address      string                 `json:"address"`
	Port         int                    `json:"port"`
	LastSeen     time.Time              `json:"last_seen"`
	Labels       map[string]string      `json:"labels,omitempty"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
	Services     map[string]Service     `json:"services,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Service represents a service running on a node
type Service struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Status   string            `json:"status"`
	Port     int               `json:"port"`
	Endpoint string            `json:"endpoint,omitempty"`
	Health   *HealthStatus     `json:"health,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// HealthStatus represents the health status of a service or node
type HealthStatus struct {
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
	URL       string    `json:"url,omitempty"`
}

// Command represents a command to be executed on a node
type Command struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`   // exec, service, config
	Target  string                 `json:"target"` // node or service name
	Action  string                 `json:"action"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Timeout time.Duration          `json:"timeout,omitempty"`
	User    string                 `json:"user,omitempty"`
	Created time.Time              `json:"created"`
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	ID       string        `json:"id"`
	NodeID   string        `json:"node_id"`
	Command  Command       `json:"command"`
	Status   string        `json:"status"` // success, error, timeout
	Output   string        `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Finished time.Time     `json:"finished"`
}

// Registry manages node registration and tracking
type Registry interface {
	Register(node *Node) error
	Unregister(id string) error
	Get(id string) (*Node, error)
	List() ([]*Node, error)
	Update(id string, updates map[string]interface{}) error
	SetStatus(id, status string) error
	GetByLabel(key, value string) ([]*Node, error)
	GetByCapability(capability string) ([]*Node, error)
}

// InMemoryRegistry provides an in-memory implementation of Registry
type InMemoryRegistry struct {
	nodes map[string]*Node
	mutex sync.RWMutex
}

// NewInMemoryRegistry creates a new in-memory node registry
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		nodes: make(map[string]*Node),
	}
}

func (r *InMemoryRegistry) Register(node *Node) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now
	node.LastSeen = now

	r.nodes[node.ID] = node
	return nil
}

func (r *InMemoryRegistry) Unregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.nodes, id)
	return nil
}

func (r *InMemoryRegistry) Get(id string) (*Node, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	node, exists := r.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return node, nil
}

func (r *InMemoryRegistry) List() ([]*Node, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	nodes := make([]*Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (r *InMemoryRegistry) Update(id string, updates map[string]interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	node, exists := r.nodes[id]
	if !exists {
		return fmt.Errorf("node not found: %s", id)
	}

	for key, value := range updates {
		switch key {
		case "status":
			node.Status = value.(string)
		case "address":
			node.Address = value.(string)
		case "port":
			node.Port = value.(int)
		case "labels":
			if m, ok := value.(map[string]string); ok {
				node.Labels = m
			}
		case "capabilities":
			if m, ok := value.(map[string]interface{}); ok {
				node.Capabilities = m
			}
		case "services":
			if m, ok := value.(map[string]Service); ok {
				node.Services = m
			}
		case "metadata":
			if m, ok := value.(map[string]interface{}); ok {
				node.Metadata = m
			}
		}
	}

	node.UpdatedAt = time.Now()
	node.LastSeen = time.Now()
	return nil
}

func (r *InMemoryRegistry) SetStatus(id, status string) error {
	return r.Update(id, map[string]interface{}{"status": status})
}

func (r *InMemoryRegistry) GetByLabel(key, value string) ([]*Node, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var nodes []*Node
	for _, node := range r.nodes {
		if node.Labels[key] == value {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (r *InMemoryRegistry) GetByCapability(capability string) ([]*Node, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var nodes []*Node
	for _, node := range r.nodes {
		if _, exists := node.Capabilities[capability]; exists {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

// Server represents the coordination server
type Server struct {
	config    *Config
	registry  Registry
	httpSrv   *http.Server
	router    *http.ServeMux
	clients   map[chan Event]bool
	clientsMu sync.Mutex
	commandCh chan CommandResult
}

// Event represents a server event for broadcasting
type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NewServer creates a new coordination server
func NewServer(cfg *Config) *Server {
	srv := &Server{
		config:    cfg,
		registry:  NewInMemoryRegistry(),
		router:    http.NewServeMux(),
		clients:   make(map[chan Event]bool),
		commandCh: make(chan CommandResult, 100),
	}

	srv.setupRoutes()
	return srv
}

func (s *Server) setupRoutes() {
	// Node management
	s.router.HandleFunc("/api/v1/nodes", s.handleNodesRequest)
	s.router.HandleFunc("/api/v1/nodes/", s.handleNodeRequest)

	// Command dispatch
	s.router.HandleFunc("/api/v1/commands/", s.handleCommandResultRequest)

	// Service discovery
	s.router.HandleFunc("/api/v1/services", s.handleListServices)

	// Health and monitoring
	s.router.HandleFunc("/health", s.handleHealth)
	s.router.HandleFunc("/metrics", s.handleMetrics)

	// WebSocket endpoint (simplified - upgrade not available in stdlib)
	s.router.HandleFunc("/ws", s.handleWebSocket)
}

// Request routing helpers
func (s *Server) handleNodesRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleRegisterNode(w, r)
	case http.MethodGet:
		s.handleListNodes(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNodeRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/nodes/")
	if path == "" {
		http.Error(w, "Node ID required", http.StatusBadRequest)
		return
	}

	parts := strings.Split(path, "/")
	nodeID := parts[0]

	// Check if this is a command request
	if len(parts) >= 3 && parts[1] == "commands" {
		switch r.Method {
		case http.MethodPost:
			s.handleSendCommand(w, r, nodeID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		if strings.HasSuffix(r.URL.Path, "/status") {
			s.handleGetNodeStatus(w, r, nodeID)
		} else {
			s.handleGetNode(w, r, nodeID)
		}
	case http.MethodPut:
		s.handleUpdateNode(w, r, nodeID)
	case http.MethodDelete:
		s.handleUnregisterNode(w, r, nodeID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNodeCommandRequest(w http.ResponseWriter, r *http.Request) {
	// This function is handled by handleNodeRequest which processes both /nodes/{id} and /nodes/{id}/commands
	http.Error(w, "Not implemented", http.StatusNotFound)
}

func (s *Server) handleCommandResultRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/commands/")
	if path == "" {
		http.Error(w, "Command ID required", http.StatusBadRequest)
		return
	}

	commandID := strings.Split(path, "/")[0]

	switch r.Method {
	case http.MethodPost:
		if strings.HasSuffix(r.URL.Path, "/result") {
			s.handleCommandResult(w, r, commandID)
		} else {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Start starts the coordination server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(s.authMiddleware(s.loggingMiddleware(s.router))),
		ReadTimeout:  s.parseTimeout(s.config.Server.ReadTimeout),
		WriteTimeout: s.parseTimeout(s.config.Server.WriteTimeout),
		IdleTimeout:  s.parseTimeout(s.config.Server.IdleTimeout),
	}

	// Start event broadcaster
	go s.broadcastResults()

	return s.httpSrv.ListenAndServe()
}

// Stop stops the coordination server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) parseTimeout(timeout string) time.Duration {
	if duration, err := time.ParseDuration(timeout); err == nil {
		return duration
	}
	return 30 * time.Second // default
}

// GetRegistry returns the node registry
func (s *Server) GetRegistry() Registry {
	return s.registry
}

func (s *Server) broadcastResults() {
	for result := range s.commandCh {
		s.broadcastEvent("command_result", result)
	}
}

func (s *Server) broadcastEvent(eventType string, data interface{}) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client, _ := range s.clients {
		select {
		case client <- event:
		default:
			close(client)
			delete(s.clients, client)
		}
	}
}

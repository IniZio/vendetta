package coordination

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// handleRegisterNode handles node registration
func (s *Server) handleRegisterNode(w http.ResponseWriter, r *http.Request) {
	var node Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.registry.Register(&node); err != nil {
		http.Error(w, fmt.Sprintf("Failed to register node: %v", err), http.StatusInternalServerError)
		return
	}

	s.broadcastEvent("node_registered", map[string]interface{}{
		"node": node,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// handleListNodes handles listing all nodes
func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.registry.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list nodes: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// handleGetNode handles getting a specific node
func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request, nodeID string) {
	node, err := s.registry.Get(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Node not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// handleGetNodeStatus handles getting node status
func (s *Server) handleGetNodeStatus(w http.ResponseWriter, r *http.Request, nodeID string) {
	node, err := s.registry.Get(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Node not found: %v", err), http.StatusNotFound)
		return
	}

	status := map[string]interface{}{
		"node_id":   node.ID,
		"status":    node.Status,
		"last_seen": node.LastSeen,
		"uptime":    time.Since(node.CreatedAt).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleUpdateNode handles updating a node
func (s *Server) handleUpdateNode(w http.ResponseWriter, r *http.Request, nodeID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.registry.Update(nodeID, updates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update node: %v", err), http.StatusInternalServerError)
		return
	}

	node, err := s.registry.Get(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve updated node: %v", err), http.StatusInternalServerError)
		return
	}

	s.broadcastEvent("node_updated", map[string]interface{}{
		"node": node,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// handleUnregisterNode handles unregistering a node
func (s *Server) handleUnregisterNode(w http.ResponseWriter, r *http.Request, nodeID string) {
	if err := s.registry.Unregister(nodeID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unregister node: %v", err), http.StatusInternalServerError)
		return
	}

	s.broadcastEvent("node_unregistered", map[string]interface{}{
		"node_id": nodeID,
	})

	w.WriteHeader(http.StatusNoContent)
}

// handleSendCommand handles sending commands to nodes
func (s *Server) handleSendCommand(w http.ResponseWriter, r *http.Request, nodeID string) {
	var command Command
	if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Generate command ID if not provided
	if command.ID == "" {
		command.ID = fmt.Sprintf("cmd_%d_%s", time.Now().Unix(), nodeID)
	}
	command.Created = time.Now()

	// Verify node exists
	_, err := s.registry.Get(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Node not found: %v", err), http.StatusNotFound)
		return
	}

	// In a real implementation, this would dispatch to the node
	// For now, we'll simulate a response
	result := CommandResult{
		ID:       command.ID,
		NodeID:   nodeID,
		Command:  command,
		Status:   "success",
		Output:   "Command executed successfully",
		Duration: time.Millisecond * 100,
		Finished: time.Now(),
	}

	// Send result through channel for broadcasting
	select {
	case s.commandCh <- result:
	default:
		log.Printf("Command result channel full, dropping result for command %s", command.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCommandResult handles receiving command results from nodes
func (s *Server) handleCommandResult(w http.ResponseWriter, r *http.Request, commandID string) {
	var result CommandResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if result.ID != commandID {
		http.Error(w, "Command ID mismatch", http.StatusBadRequest)
		return
	}

	// Send result through channel for broadcasting
	select {
	case s.commandCh <- result:
	default:
		log.Printf("Command result channel full, dropping result for command %s", commandID)
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleListServices handles listing all services across all nodes
func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.registry.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list nodes: %v", err), http.StatusInternalServerError)
		return
	}

	services := make(map[string][]Service)
	for _, node := range nodes {
		if len(node.Services) > 0 {
			services[node.ID] = make([]Service, 0, len(node.Services))
			for _, service := range node.Services {
				services[node.ID] = append(services[node.ID], service)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
		"nodes":    len(nodes),
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.registry.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	activeNodes := 0
	for _, node := range nodes {
		if node.Status == "active" {
			activeNodes++
		}
	}

	health := map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now(),
		"total_nodes":  len(nodes),
		"active_nodes": activeNodes,
		"version":      "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleMetrics handles metrics requests
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.registry.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list nodes: %v", err), http.StatusInternalServerError)
		return
	}

	statusCounts := make(map[string]int)
	providerCounts := make(map[string]int)
	var totalServices int

	for _, node := range nodes {
		statusCounts[node.Status]++
		providerCounts[node.Provider]++
		totalServices += len(node.Services)
	}

	s.clientsMu.Lock()
	connectedClients := len(s.clients)
	s.clientsMu.Unlock()

	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"nodes": map[string]interface{}{
			"total":       len(nodes),
			"by_status":   statusCounts,
			"by_provider": providerCounts,
		},
		"services": map[string]interface{}{
			"total": totalServices,
		},
		"websocket": map[string]interface{}{
			"connected_clients": connectedClients,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleWebSocket handles WebSocket connections for real-time updates
// Simplified version using Server-Sent Events since WebSocket upgrade requires gorilla/websocket
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "websocket" {
		// For WebSocket support, gorilla/websocket would be needed
		// For now, we'll use Server-Sent Events as fallback
		s.handleServerSentEvents(w, r)
		return
	}

	// Server-Sent Events endpoint
	s.handleServerSentEvents(w, r)
}

// handleServerSentEvents provides SSE streaming for real-time updates
func (s *Server) handleServerSentEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create client channel
	clientChan := make(chan Event, 10)

	s.clientsMu.Lock()
	s.clients[clientChan] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, clientChan)
		s.clientsMu.Unlock()
		close(clientChan)
	}()

	// Send initial state
	nodes, _ := s.registry.List()
	initialEvent := Event{
		Type:      "initial_state",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"nodes": nodes,
		},
	}

	eventData, _ := json.Marshal(initialEvent)
	fmt.Fprintf(w, "data: %s\n\n", eventData)
	flusher.Flush()

	// Stream events
	for {
		select {
		case event := <-clientChan:
			eventData, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", eventData)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// Middleware functions

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	if !s.config.Auth.Enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Simple Bearer token validation
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != s.config.Server.AuthToken && token != s.config.Auth.JWTSecret {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter is a wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vibegear/vendatta/pkg/coordination"
)

func main() {
	baseURL := "http://localhost:3001"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	client := NewCoordinationClient(baseURL)

	// Demo: Register a node
	fmt.Println("Registering test node...")
	node := &coordination.Node{
		ID:       "test-node-client",
		Name:     "Test Client Node",
		Provider: "docker",
		Status:   "active",
		Address:  "localhost",
		Port:     8080,
		Labels: map[string]string{
			"env":  "test",
			"type": "client",
		},
		Capabilities: map[string]interface{}{
			"docker": true,
			"test":   true,
		},
		Services: map[string]coordination.Service{
			"web": {
				ID:       "web-service",
				Name:     "Web Server",
				Type:     "http",
				Status:   "running",
				Port:     8080,
				Endpoint: "http://localhost:8080",
			},
		},
	}

	if err := client.RegisterNode(node); err != nil {
		log.Fatalf("Failed to register node: %v", err)
	}
	fmt.Println("✅ Node registered successfully")

	// Demo: List all nodes
	fmt.Println("\nListing all nodes...")
	nodes, err := client.ListNodes()
	if err != nil {
		log.Fatalf("Failed to list nodes: %v", err)
	}
	fmt.Printf("Found %d nodes:\n", len(nodes))
	for _, n := range nodes {
		fmt.Printf("  - %s (%s) [%s] at %s:%d\n", n.ID, n.Name, n.Status, n.Address, n.Port)
	}

	// Demo: Send a command
	fmt.Println("\nSending test command...")
	result, err := client.SendCommand("test-node-client", &coordination.Command{
		Type:   "exec",
		Action: "echo 'Hello from coordination client!'",
		Params: map[string]interface{}{
			"timeout": "10s",
		},
	})
	if err != nil {
		log.Fatalf("Failed to send command: %v", err)
	}
	fmt.Printf("✅ Command sent: %s\n", result.ID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Output: %s\n", result.Output)

	// Demo: Check health
	fmt.Println("\nChecking server health...")
	health, err := client.CheckHealth()
	if err != nil {
		log.Fatalf("Failed to check health: %v", err)
	}
	fmt.Printf("Server health: %s\n", health["status"])
	fmt.Printf("Total nodes: %v\n", health["total_nodes"])
	fmt.Printf("Active nodes: %v\n", health["active_nodes"])

	// Demo: Listen for real-time updates
	fmt.Println("\nStarting real-time event listener...")
	fmt.Println("Press Ctrl+C to exit...")

	events, err := client.SubscribeEvents()
	if err != nil {
		log.Fatalf("Failed to subscribe to events: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for event := range events {
			fmt.Printf("\n🔔 Event: %s\n", event.Type)
			if event.Type == "node_registered" {
				if nodeData, ok := event.Data.(map[string]interface{})["node"]; ok {
					nodeJSON, _ := json.MarshalIndent(nodeData, "", "  ")
					fmt.Printf("Node details:\n%s\n", nodeJSON)
				}
			}
		}
	}()

	<-sigCh
	fmt.Println("\n👋 Shutting down client...")
}

// CoordinationClient provides a simple client for the coordination server
type CoordinationClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCoordinationClient(baseURL string) *CoordinationClient {
	return &CoordinationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *CoordinationClient) RegisterNode(node *coordination.Node) error {
	data, err := json.Marshal(node)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/nodes",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *CoordinationClient) ListNodes() ([]*coordination.Node, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/nodes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Nodes []*coordination.Node `json:"nodes"`
		Count int                  `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Nodes, nil
}

func (c *CoordinationClient) SendCommand(nodeID string, command *coordination.Command) (*coordination.CommandResult, error) {
	data, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/nodes/"+nodeID+"/commands",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("command failed with status: %d", resp.StatusCode)
	}

	var result coordination.CommandResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *CoordinationClient) CheckHealth() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, err
	}

	return health, nil
}

func (c *CoordinationClient) SubscribeEvents() (<-chan coordination.Event, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/ws")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to connect to events: %d", resp.StatusCode)
	}

	events := make(chan coordination.Event, 100)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		decoder := json.NewDecoder(resp.Body)

		for {
			var event coordination.Event
			if err := decoder.Decode(&event); err != nil {
				if err.Error() != "EOF" {
					log.Printf("Error decoding event: %v", err)
				}
				return
			}

			select {
			case events <- event:
			default:
				log.Printf("Event channel full, dropping event: %s", event.Type)
			}
		}
	}()

	return events, nil
}

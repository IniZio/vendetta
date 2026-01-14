package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibegear/vendetta/pkg/coordination"
	"github.com/vibegear/vendetta/pkg/transport"
)

type LocalTransportTestSuite struct {
	testEnv     *TestEnvironment
	tempDir     string
	coordConfig string
}

func (s *LocalTransportTestSuite) SetupTestSuite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vendetta-local-transport-*")
	require.NoError(t, err)
	s.tempDir = tempDir
	s.coordConfig = filepath.Join(tempDir, "coordination.yaml")
}

func (s *LocalTransportTestSuite) TeardownTestSuite(t *testing.T) {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func TestLocalTransportManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping local transport test in short mode")
	}

	s := &LocalTransportTestSuite{}
	s.SetupTestSuite(t)
	defer s.TeardownTestSuite(t)

	t.Run("create_ssh_transport_for_localhost", func(t *testing.T) {
		cfg := transport.CreateDefaultSSHConfig(
			"localhost:22",
			"root",
			"/dev/null",
		)
		cfg.Auth.KeyData = []byte{}

		manager := transport.NewManager()
		err := manager.RegisterConfig("local-ssh", cfg)
		require.NoError(t, err)

		tr, err := manager.CreateTransport("local-ssh")
		require.NoError(t, err)
		require.NotNil(t, tr)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = tr.Connect(ctx, "localhost:22")
		if err != nil {
			t.Logf("Expected connection error: %v", err)
		}

		tr.Disconnect(ctx)
	})

	t.Run("create_http_transport_local", func(t *testing.T) {
		cfg := transport.CreateDefaultHTTPConfig(
			"http://localhost:3001",
			"test-token",
		)

		manager := transport.NewManager()
		err := manager.RegisterConfig("local-http", cfg)
		require.NoError(t, err)

		tr, err := manager.CreateTransport("local-http")
		require.NoError(t, err)
		require.NotNil(t, tr)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = tr.Connect(ctx, "http://localhost:3001")
		if err != nil {
			t.Logf("Expected connection error: %v", err)
		}

		tr.Disconnect(ctx)
	})

	runTransportPoolTests(t)
}

func runTransportPoolTests(t *testing.T) {
	t.Run("connection_pooling", func(t *testing.T) {
		cfg := transport.CreateDefaultSSHConfig(
			"localhost:22",
			"root",
			"/dev/null",
		)
		cfg.Auth.KeyData = []byte{}

		manager := transport.NewManager()
		err := manager.RegisterConfig("pool-test", cfg)
		require.NoError(t, err)

		pool, err := manager.CreatePool("pool-test")
		require.NoError(t, err)
		require.NotNil(t, pool)
		defer pool.Close()

		metrics := pool.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, 0, metrics.Active)
		assert.Equal(t, 0, metrics.Idle)
	})

	t.Run("pool_connection_limits", func(t *testing.T) {
		cfg := transport.CreateDefaultSSHConfig(
			"limits-test:22",
			"root",
			"/dev/null",
		)
		cfg.Auth.KeyData = []byte{}
		cfg.Connection.MaxConns = 2
		cfg.Connection.MaxIdle = 1

		manager := transport.NewManager()
		err := manager.RegisterConfig("limits-test", cfg)
		require.NoError(t, err)

		pool, err := manager.CreatePool("limits-test")
		require.NoError(t, err)
		defer pool.Close()

		metrics := pool.GetMetrics()
		assert.NotNil(t, metrics)
	})
}

func TestLocalCoordinationServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping coordination server test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "vendetta-coord-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "coordination.yaml")

	err = coordination.GenerateDefaultConfig(configPath)
	require.NoError(t, err)

	cfg, err := coordination.LoadConfig(configPath)
	require.NoError(t, err)

	err = coordination.ValidateConfig(cfg)
	require.NoError(t, err)

	err = coordination.CheckPortAvailable(cfg.Server.Host, cfg.Server.Port)
	if err != nil {
		t.Logf("Port %d may be in use: %v", cfg.Server.Port, err)
		cfg.Server.Port = 3002
	}

	t.Run("server_creation", func(t *testing.T) {
		srv := coordination.NewServer(cfg)
		require.NotNil(t, srv)

		registry := srv.GetRegistry()
		require.NotNil(t, registry)
	})

	t.Run("registry_operations", func(t *testing.T) {
		registry := coordination.NewInMemoryRegistry()

		node := &coordination.Node{
			ID:       "test-node-1",
			Name:     "test-node",
			Provider: "docker",
			Status:   "active",
		}
		err := registry.Register(node)
		require.NoError(t, err)

		nodes, err := registry.List()
		require.NoError(t, err)
		assert.Len(t, nodes, 1)

		retrieved, err := registry.Get("test-node-1")
		require.NoError(t, err)
		assert.Equal(t, "test-node-1", retrieved.ID)

		err = registry.Update("test-node-1", map[string]interface{}{"status": "busy"})
		require.NoError(t, err)

		retrieved, err = registry.Get("test-node-1")
		require.NoError(t, err)
		assert.Equal(t, "busy", retrieved.Status)

		err = registry.Unregister("test-node-1")
		require.NoError(t, err)

		_, err = registry.Get("test-node-1")
		assert.Error(t, err)
	})

	t.Run("node_query_by_label", func(t *testing.T) {
		registry := coordination.NewInMemoryRegistry()

		node1 := &coordination.Node{
			ID:     "node-1",
			Name:   "node-1",
			Labels: map[string]string{"env": "test"},
		}
		node2 := &coordination.Node{
			ID:     "node-2",
			Name:   "node-2",
			Labels: map[string]string{"env": "prod"},
		}
		node3 := &coordination.Node{
			ID:     "node-3",
			Name:   "node-3",
			Labels: map[string]string{"env": "test", "type": "gpu"},
		}

		require.NoError(t, registry.Register(node1))
		require.NoError(t, registry.Register(node2))
		require.NoError(t, registry.Register(node3))

		testNodes, err := registry.GetByLabel("env", "test")
		require.NoError(t, err)
		assert.Len(t, testNodes, 2)
	})

	t.Run("node_query_by_capability", func(t *testing.T) {
		registry := coordination.NewInMemoryRegistry()

		node1 := &coordination.Node{
			ID:           "node-1",
			Name:         "node-1",
			Capabilities: map[string]interface{}{"docker": true, "gpu": true},
		}
		node2 := &coordination.Node{
			ID:           "node-2",
			Name:         "node-2",
			Capabilities: map[string]interface{}{"docker": true},
		}

		require.NoError(t, registry.Register(node1))
		require.NoError(t, registry.Register(node2))

		gpuNodes, err := registry.GetByCapability("gpu")
		require.NoError(t, err)
		assert.Len(t, gpuNodes, 1)
		assert.Equal(t, "node-1", gpuNodes[0].ID)
	})
}

func TestLocalTransportExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping transport execution test in short mode")
	}

	t.Run("local_command_execution", func(t *testing.T) {
		cmd := &transport.Command{
			Cmd:           []string{"echo", "hello"},
			CaptureOutput: true,
			Timeout:       5 * time.Second,
		}

		assert.NotNil(t, cmd)
		assert.Equal(t, []string{"echo", "hello"}, cmd.Cmd)
		assert.True(t, cmd.CaptureOutput)
	})

	t.Run("command_result_structure", func(t *testing.T) {
		result := &transport.Result{
			ExitCode: 0,
			Output:   "test output",
			Duration: 100 * time.Millisecond,
		}

		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, "test output", result.Output)
		assert.Greater(t, result.Duration.Milliseconds(), int64(0))
	})

	t.Run("transport_error_types", func(t *testing.T) {
		errors := []error{
			transport.ErrNotConnected,
			transport.ErrInvalidTarget,
			transport.ErrAuthFailed,
			transport.ErrTimeout,
			transport.ErrConnectionFailed,
		}

		for _, err := range errors {
			assert.Error(t, err)
			assert.NotEmpty(t, err.Error())
		}

		assert.True(t, transport.ErrTimeout.Retryable)
		assert.True(t, transport.ErrConnectionFailed.Retryable)
		assert.False(t, transport.ErrAuthFailed.Retryable)
	})
}

func TestLocalTransportHTTPAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping HTTP transport test in short mode")
	}

	t.Run("http_client_configuration", func(t *testing.T) {
		cfg := transport.CreateDefaultHTTPConfig(
			"http://localhost:3001",
			"test-token",
		)

		assert.Equal(t, "http", cfg.Protocol)
		assert.Equal(t, "http://localhost:3001", cfg.Target)
		assert.Equal(t, "token", cfg.Auth.Type)
		assert.Equal(t, "test-token", cfg.Auth.Token)
	})

	t.Run("http_auth_headers", func(t *testing.T) {
		cfg := transport.CreateDefaultHTTPConfig(
			"https://api.example.com",
			"secret-token",
		)

		manager := transport.NewManager()
		err := manager.RegisterConfig("http-test", cfg)
		require.NoError(t, err)

		tr, err := manager.CreateTransport("http-test")
		require.NoError(t, err)
		require.NotNil(t, tr)

		info := tr.GetInfo()
		assert.NotNil(t, info)
		assert.Equal(t, "http", info.Protocol)
	})

	t.Run("http_security_config", func(t *testing.T) {
		cfg := transport.CreateDefaultHTTPConfig(
			"https://secure.example.com",
			"token",
		)
		cfg.Security.SkipTLSVerify = false
		cfg.Security.VerifyCertificate = true
		cfg.Security.CACertPath = "/path/to/ca.crt"

		assert.False(t, cfg.Security.SkipTLSVerify)
		assert.True(t, cfg.Security.VerifyCertificate)
		assert.Equal(t, "/path/to/ca.crt", cfg.Security.CACertPath)
	})
}

func TestLocalTransportIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("manager_lifecycle", func(t *testing.T) {
		manager := transport.NewManager()

		sshCfg := transport.CreateDefaultSSHConfig(
			"host1:22",
			"user1",
			"/path/to/key1",
		)
		httpCfg := transport.CreateDefaultHTTPConfig(
			"http://api1.example.com",
			"token1",
		)

		err := manager.RegisterConfig("ssh-1", sshCfg)
		require.NoError(t, err)
		err = manager.RegisterConfig("http-1", httpCfg)
		require.NoError(t, err)

		configs := manager.ListConfigs()
		require.NotNil(t, configs)
		assert.GreaterOrEqual(t, len(configs), 2)

		sshTr, err := manager.CreateTransport("ssh-1")
		require.NoError(t, err)
		assert.NotNil(t, sshTr)

		httpTr, err := manager.CreateTransport("http-1")
		require.NoError(t, err)
		assert.NotNil(t, httpTr)

		sshInfo := sshTr.GetInfo()
		assert.Equal(t, "ssh", sshInfo.Protocol)
		assert.Equal(t, "host1:22", sshInfo.Target)
		assert.False(t, sshInfo.Connected)

		httpInfo := httpTr.GetInfo()
		assert.Equal(t, "http", httpInfo.Protocol)
		assert.Equal(t, "http://api1.example.com", httpInfo.Target)
		assert.False(t, httpInfo.Connected)

		savePath := filepath.Join(t.TempDir(), "transports.yaml")
		err = manager.SaveConfig(savePath)
		require.NoError(t, err)

		manager2 := transport.NewManager()
		err = manager2.LoadConfig(savePath)
		require.NoError(t, err)

		manager.CloseAll()
		manager2.CloseAll()
	})

	t.Run("transport_factory", func(t *testing.T) {
		manager := transport.NewManager()

		sshCfg := transport.CreateDefaultSSHConfig(
			"test:22",
			"user",
			"/dev/null",
		)
		sshCfg.Auth.KeyData = []byte{}

		err := manager.RegisterConfig("factory-test", sshCfg)
		require.NoError(t, err)

		tr, err := manager.CreateTransport("factory-test")
		require.NoError(t, err)
		assert.NotNil(t, tr)

		info := tr.GetInfo()
		assert.Equal(t, "ssh", info.Protocol)

		manager.CloseAll()
	})
}

func BenchmarkLocalTransport(b *testing.B) {
	b.Run("transport_creation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cfg := transport.CreateDefaultSSHConfig(
				"localhost:22",
				"root",
				"/dev/null",
			)
			cfg.Auth.KeyData = []byte{}

			manager := transport.NewManager()
			manager.RegisterConfig("bench", cfg)
			manager.CloseAll()
		}
	})

	b.Run("command_serialization", func(b *testing.B) {
		cmd := &transport.Command{
			Cmd:           []string{"ls", "-la", "/tmp"},
			CaptureOutput: true,
			Timeout:       30 * time.Second,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(cmd)
			var parsed transport.Command
			json.Unmarshal(data, &parsed)
		}
	})
}

func makeCoordRequest(method, url string, body interface{}) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp, respBody, nil
}

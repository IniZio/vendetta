package agent

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vibegear/oursky/pkg/provider"
)

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Create(ctx context.Context, sessionID string, workspacePath string, config interface{}) (*provider.Session, error) {
	args := m.Called(ctx, sessionID, workspacePath, config)
	return args.Get(0).(*provider.Session), args.Error(1)
}

func (m *MockProvider) Start(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockProvider) Stop(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockProvider) Destroy(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockProvider) Exec(ctx context.Context, sessionID string, opts provider.ExecOptions) error {
	args := m.Called(ctx, sessionID, opts)
	return args.Error(0)
}

func (m *MockProvider) List(ctx context.Context) ([]provider.Session, error) {
	args := m.Called(ctx)
	return args.Get(0).([]provider.Session), args.Error(1)
}

func TestAgentServer_HandleExec(t *testing.T) {
	mockP := new(MockProvider)
	server := NewAgentServer("test-session", mockP)

	req := mcp.CallToolRequest{}
	req.Params.Name = "exec"
	req.Params.Arguments = map[string]interface{}{
		"cmd": "echo hello",
	}

	mockP.On("Exec", mock.Anything, "test-session", provider.ExecOptions{
		Cmd:    []string{"/bin/bash", "-c", "echo hello"},
		Stdout: true,
		Stderr: true,
	}).Return(nil)

	result, err := server.handleExec(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "successfully")

	mockP.AssertExpectations(t)
}

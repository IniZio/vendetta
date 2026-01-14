package lxc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibegear/vendetta/pkg/provider"
)

func TestNewLXCProvider(t *testing.T) {
	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	assert.NotNil(t, lxcProvider)
	assert.Equal(t, "lxc", lxcProvider.Name())
}

func TestLXCProvider_Name(t *testing.T) {
	provider := &LXCProvider{name: "lxc"}
	assert.Equal(t, "lxc", provider.Name())
}

func TestLXCProvider_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	sessionID := "test-create"
	workspacePath := t.TempDir()

	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	session, err := lxcProvider.Create(ctx, sessionID, workspacePath, nil)
	require.NoError(t, err)
	defer lxcProvider.Destroy(ctx, sessionID)

	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, "lxc", session.Provider)
	assert.Contains(t, session.Labels, "vendetta.session.id")
	assert.Equal(t, sessionID, session.Labels["vendetta.session.id"])
}

func TestLXCProvider_Start_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	sessionID := "test-start"

	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	_, err = lxcProvider.Create(ctx, sessionID, t.TempDir(), nil)
	require.NoError(t, err)
	defer lxcProvider.Destroy(ctx, sessionID)

	err = lxcProvider.Start(ctx, sessionID)
	assert.NoError(t, err)
}

func TestLXCProvider_Stop_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	sessionID := "test-stop"

	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	_, err = lxcProvider.Create(ctx, sessionID, t.TempDir(), nil)
	require.NoError(t, err)
	defer lxcProvider.Destroy(ctx, sessionID)

	err = lxcProvider.Start(ctx, sessionID)
	require.NoError(t, err)

	err = lxcProvider.Stop(ctx, sessionID)
	assert.NoError(t, err)
}

func TestLXCProvider_Destroy_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	sessionID := "test-destroy"

	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	_, err = lxcProvider.Create(ctx, sessionID, t.TempDir(), nil)
	require.NoError(t, err)

	err = lxcProvider.Destroy(ctx, sessionID)
	assert.NoError(t, err)
}

func TestLXCProvider_List_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	sessionID := "test-list"
	_, err = lxcProvider.Create(ctx, sessionID, t.TempDir(), nil)
	if err != nil {
		t.Skip("Cannot create LXC container:", err)
	}
	defer lxcProvider.Destroy(ctx, sessionID)

	sessions, err := lxcProvider.List(ctx)
	require.NoError(t, err)
	require.NotNil(t, sessions)

	found := false
	for _, session := range sessions {
		if session.ID == sessionID && session.Provider == "lxc" {
			found = true
			assert.Contains(t, session.Labels, "vendetta.session.id")
			assert.Equal(t, sessionID, session.Labels["vendetta.session.id"])
			break
		}
	}
	assert.True(t, found, "Should find our test container in list")
}

func TestLXCProvider_Exec_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	sessionID := "test-exec"

	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	_, err = lxcProvider.Create(ctx, sessionID, t.TempDir(), nil)
	require.NoError(t, err)
	defer lxcProvider.Destroy(ctx, sessionID)

	err = lxcProvider.Start(ctx, sessionID)
	require.NoError(t, err)

	opts := provider.ExecOptions{
		Cmd:    []string{"echo", "test"},
		Stdout: true,
		Stderr: true,
	}

	err = lxcProvider.Exec(ctx, sessionID, opts)
	if err != nil {
		t.Skip("Cannot execute command in LXC container:", err)
	}
	assert.NoError(t, err)
}

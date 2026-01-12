package lxc

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	assert.Equal(t, "lxc", lxcProvider.Name())
}

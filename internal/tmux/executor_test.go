package tmux_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestRealExecutor_TmuxNotRunning(t *testing.T) {
	exec := tmux.NewRealExecutor()
	_, err := exec.Execute("list-sessions")
	// Allow failure (tmux might not be running), but should not panic
	_ = err
}

func TestTmuxAvailable(t *testing.T) {
	_, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux not installed")
	}
	assert.NoError(t, err)
}

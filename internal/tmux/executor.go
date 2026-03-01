package tmux

import (
	"os/exec"
	"strings"
)

// RealExecutor 透過 os/exec 呼叫實際的 tmux 指令。
type RealExecutor struct{}

// NewRealExecutor 建立 RealExecutor。
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Execute 執行 tmux 指令並回傳輸出結果。
func (e *RealExecutor) Execute(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

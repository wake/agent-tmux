# TSM (tmux-session-menu) Phase 1 實作計畫

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 建立一個輕量但資訊豐富的 tmux session 選單 TUI，支援群組、排序、三層狀態偵測、底部預覽。

**Architecture:** Go + Bubble Tea (Elm Architecture) TUI。tmux 互動層封裝所有 tmux 指令操作，三層狀態偵測（hooks → pane title → terminal content）提供 AI session 狀態。SQLite (WAL mode) 儲存群組和排序。TOML 設定檔。

**Tech Stack:** Go 1.24+, Bubble Tea, Lipgloss, Bubbles, BurntSushi/toml, modernc.org/sqlite, testify

**Design Doc:** `docs/plans/2026-03-01-tsm-design.md`

---

## Task 1: 專案腳手架

建立 Go module、目錄結構、基本 CI。

**Files:**
- Create: `go.mod`
- Create: `cmd/tsm/main.go`
- Create: `Makefile`
- Create: `.gitignore`

**Step 1: 初始化 Go module**

```bash
cd /Library/Workspace/github/wake/tmux-session-menu
go mod init github.com/wake/tmux-session-menu
```

**Step 2: 建立目錄結構**

```bash
mkdir -p cmd/tsm
mkdir -p internal/{ui,tmux,config,store,ai}
```

**Step 3: 建立 main.go 佔位**

Create `cmd/tsm/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("tsm - tmux session menu")
}
```

**Step 4: 建立 .gitignore**

Create `.gitignore`:
```
# Binary
tsm
/cmd/tsm/tsm

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store

# Test
coverage.out

# Build
dist/
```

**Step 5: 建立 Makefile**

Create `Makefile`:
```makefile
.PHONY: build test lint clean

BINARY=tsm

build:
	go build -o $(BINARY) ./cmd/tsm

test:
	go test ./... -v -race

test-cover:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -html=coverage.out

lint:
	go vet ./...

clean:
	rm -f $(BINARY) coverage.out

run: build
	./$(BINARY)
```

**Step 6: 驗證編譯**

```bash
go build ./cmd/tsm
```
Expected: 成功，產生 `tsm` 二進位。

**Step 7: Commit**

```bash
git add go.mod cmd/ Makefile .gitignore internal/
git commit -m "chore: 初始化 Go 專案結構"
```

---

## Task 2: Config 模組 — TOML 載入

設定檔載入、預設值、路徑解析。

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: 安裝依賴**

```bash
go get github.com/BurntSushi/toml
```

**Step 2: 寫失敗的測試 — 預設設定**

Create `internal/config/config_test.go`:
```go
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wake/tmux-session-menu/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()

	assert.Equal(t, "~/.config/tsm", cfg.DataDir)
	assert.Equal(t, 150, cfg.PreviewLines)
	assert.Equal(t, 2, cfg.PollIntervalSec)
}
```

**Step 3: 執行測試確認失敗**

```bash
go get github.com/stretchr/testify
go test ./internal/config/ -v -run TestDefaultConfig
```
Expected: FAIL — `config.Default` 未定義。

**Step 4: 實作最小程式碼讓測試通過**

Create `internal/config/config.go`:
```go
package config

// Config 是 TSM 的全域設定。
type Config struct {
	DataDir         string `toml:"data_dir"`
	PreviewLines    int    `toml:"preview_lines"`
	PollIntervalSec int    `toml:"poll_interval_sec"`
}

// Default 回傳預設設定。
func Default() Config {
	return Config{
		DataDir:         "~/.config/tsm",
		PreviewLines:    150,
		PollIntervalSec: 2,
	}
}
```

**Step 5: 執行測試確認通過**

```bash
go test ./internal/config/ -v -run TestDefaultConfig
```
Expected: PASS

**Step 6: 寫失敗的測試 — 從 TOML 字串載入**

在 `internal/config/config_test.go` 追加：
```go
func TestLoadFromTOML(t *testing.T) {
	tomlData := `
data_dir = "/tmp/tsm-test"
preview_lines = 50
poll_interval_sec = 5
`
	cfg, err := config.LoadFromString(tomlData)
	require.NoError(t, err)

	assert.Equal(t, "/tmp/tsm-test", cfg.DataDir)
	assert.Equal(t, 50, cfg.PreviewLines)
	assert.Equal(t, 5, cfg.PollIntervalSec)
}

func TestLoadFromTOML_PartialOverride(t *testing.T) {
	tomlData := `preview_lines = 80`

	cfg, err := config.LoadFromString(tomlData)
	require.NoError(t, err)

	// 未指定欄位使用預設值
	assert.Equal(t, "~/.config/tsm", cfg.DataDir)
	assert.Equal(t, 80, cfg.PreviewLines)
	assert.Equal(t, 2, cfg.PollIntervalSec)
}
```

**Step 7: 執行測試確認失敗**

```bash
go test ./internal/config/ -v -run TestLoadFromTOML
```
Expected: FAIL — `config.LoadFromString` 未定義。

**Step 8: 實作 LoadFromString**

在 `internal/config/config.go` 追加：
```go
import "github.com/BurntSushi/toml"

// LoadFromString 從 TOML 字串載入設定，未指定欄位使用預設值。
func LoadFromString(data string) (Config, error) {
	cfg := Default()
	if _, err := toml.Decode(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
```

**Step 9: 執行測試確認通過**

```bash
go test ./internal/config/ -v
```
Expected: 全部 PASS

**Step 10: 寫失敗的測試 — ExpandPath**

在 `internal/config/config_test.go` 追加：
```go
import "os"

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/.config/tsm", home + "/.config/tsm"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := config.ExpandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

**Step 11: 執行測試確認失敗**

```bash
go test ./internal/config/ -v -run TestExpandPath
```
Expected: FAIL

**Step 12: 實作 ExpandPath**

在 `internal/config/config.go` 追加：
```go
import (
	"os"
	"strings"
)

// ExpandPath 將 ~ 展開為使用者家目錄。
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:]
	}
	return path
}
```

**Step 13: 執行全部測試確認通過**

```bash
go test ./internal/config/ -v
```
Expected: 全部 PASS

**Step 14: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "feat(config): TOML 設定載入，預設值與路徑展開"
```

---

## Task 3: Tmux Session 層 — 資料模型與列表

封裝 tmux 操作，提供 Session 資料結構和列表功能。此 task 使用 interface 以支援測試。

**Files:**
- Create: `internal/tmux/types.go`
- Create: `internal/tmux/session.go`
- Create: `internal/tmux/session_test.go`

**Step 1: 寫失敗的測試 — Session 資料結構與時間格式**

Create `internal/tmux/types.go`:
```go
package tmux

import "time"

// SessionStatus 表示 session 的狀態。
type SessionStatus int

const (
	StatusIdle    SessionStatus = iota // 閒置
	StatusRunning                      // AI 正在工作
	StatusWaiting                      // 等待人類輸入
	StatusError                        // 錯誤
)

// Session 代表一個 tmux session。
type Session struct {
	Name       string
	ID         string // tmux session id
	Path       string // 工作目錄
	Attached   bool
	Activity   time.Time // 最後活動時間
	Status     SessionStatus
	AIModel    string // 偵測到的 AI 模型（空字串表示非 AI session）
	AISummary  string // AI 摘要
	GroupName  string // 所屬群組
	SortOrder  int    // 排序順序
}

// Executor 定義 tmux 指令的執行介面（方便測試 mock）。
type Executor interface {
	Execute(args ...string) (string, error)
}
```

Create `internal/tmux/session_test.go`:
```go
package tmux_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestSession_RelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		activity time.Time
		expected string
	}{
		{"just now", time.Now().Add(-30 * time.Second), "30s"},
		{"minutes", time.Now().Add(-5 * time.Minute), "5m"},
		{"hours", time.Now().Add(-3 * time.Hour), "3h"},
		{"days", time.Now().Add(-2 * 24 * time.Hour), "2d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tmux.Session{Activity: tt.activity}
			assert.Equal(t, tt.expected, s.RelativeTime())
		})
	}
}

func TestSession_StatusIcon(t *testing.T) {
	tests := []struct {
		status   tmux.SessionStatus
		expected string
	}{
		{tmux.StatusRunning, "●"},
		{tmux.StatusWaiting, "◐"},
		{tmux.StatusIdle, "○"},
		{tmux.StatusError, "✗"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			s := tmux.Session{Status: tt.status}
			assert.Equal(t, tt.expected, s.StatusIcon())
		})
	}
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run "TestSession_RelativeTime|TestSession_StatusIcon"
```
Expected: FAIL — `RelativeTime` 和 `StatusIcon` 未定義。

**Step 3: 實作 Session 方法**

在 `internal/tmux/types.go` 追加：
```go
import "fmt"

// RelativeTime 回傳距今的相對時間字串。
func (s Session) RelativeTime() string {
	d := time.Since(s.Activity)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// StatusIcon 回傳狀態對應的 Unicode 圖示。
func (s Session) StatusIcon() string {
	switch s.Status {
	case StatusRunning:
		return "●"
	case StatusWaiting:
		return "◐"
	case StatusIdle:
		return "○"
	case StatusError:
		return "✗"
	default:
		return "○"
	}
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/tmux/ -v -run "TestSession_RelativeTime|TestSession_StatusIcon"
```
Expected: PASS

**Step 5: 寫失敗的測試 — 解析 tmux list-sessions 輸出**

在 `internal/tmux/session_test.go` 追加：
```go
func TestParseListSessions(t *testing.T) {
	// 模擬 tmux list-sessions -F 的輸出格式
	output := `my-project:$1:1:/home/user/project:1:1709312400
api-server:$2:0:/home/user/api:0:1709308800`

	sessions, err := tmux.ParseListSessions(output)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)

	assert.Equal(t, "my-project", sessions[0].Name)
	assert.Equal(t, "$1", sessions[0].ID)
	assert.Equal(t, "/home/user/project", sessions[0].Path)
	assert.True(t, sessions[0].Attached)

	assert.Equal(t, "api-server", sessions[1].Name)
	assert.Equal(t, "$2", sessions[1].ID)
	assert.Equal(t, "/home/user/api", sessions[1].Path)
	assert.False(t, sessions[1].Attached)
}

func TestParseListSessions_Empty(t *testing.T) {
	sessions, err := tmux.ParseListSessions("")
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}
```

**Step 6: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestParseListSessions
```
Expected: FAIL

**Step 7: 實作 ParseListSessions**

Create `internal/tmux/session.go`:
```go
package tmux

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ListSessionsFormat 是 tmux list-sessions -F 的格式字串。
const ListSessionsFormat = "#{session_name}:#{session_id}:#{session_windows}:#{session_path}:#{session_attached}:#{session_activity}"

// ParseListSessions 解析 tmux list-sessions 的輸出。
func ParseListSessions(output string) ([]Session, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	sessions := make([]Session, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// format: name:id:windows:path:attached:activity
		parts := strings.SplitN(line, ":", 6)
		if len(parts) < 6 {
			return nil, fmt.Errorf("unexpected format: %q", line)
		}

		attached := parts[4] == "1"
		activityUnix, err := strconv.ParseInt(parts[5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid activity timestamp %q: %w", parts[5], err)
		}

		sessions = append(sessions, Session{
			Name:     parts[0],
			ID:       parts[1],
			Path:     parts[3],
			Attached: attached,
			Activity: time.Unix(activityUnix, 0),
		})
	}

	return sessions, nil
}
```

**Step 8: 執行測試確認通過**

```bash
go test ./internal/tmux/ -v
```
Expected: 全部 PASS

**Step 9: 寫失敗的測試 — Manager 介面 (使用 mock executor)**

在 `internal/tmux/session_test.go` 追加：
```go
// mockExecutor 用於測試，模擬 tmux 指令。
type mockExecutor struct {
	outputs map[string]string
	err     error
}

func (m *mockExecutor) Execute(args ...string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	key := strings.Join(args, " ")
	return m.outputs[key], nil
}

func TestManager_ListSessions(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{
			"list-sessions -F #{session_name}:#{session_id}:#{session_windows}:#{session_path}:#{session_attached}:#{session_activity}": `work:$1:1:/home/user/work:1:1709312400`,
		},
	}

	mgr := tmux.NewManager(mock)
	sessions, err := mgr.ListSessions()

	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "work", sessions[0].Name)
}

func TestManager_KillSession(t *testing.T) {
	mock := &mockExecutor{outputs: map[string]string{
		"kill-session -t my-session": "",
	}}

	mgr := tmux.NewManager(mock)
	err := mgr.KillSession("my-session")
	assert.NoError(t, err)
}

func TestManager_RenameSession(t *testing.T) {
	mock := &mockExecutor{outputs: map[string]string{
		"rename-session -t old-name new-name": "",
	}}

	mgr := tmux.NewManager(mock)
	err := mgr.RenameSession("old-name", "new-name")
	assert.NoError(t, err)
}

func TestManager_NewSession(t *testing.T) {
	mock := &mockExecutor{outputs: map[string]string{
		"new-session -d -s test-session -c /tmp": "",
	}}

	mgr := tmux.NewManager(mock)
	err := mgr.NewSession("test-session", "/tmp")
	assert.NoError(t, err)
}

func TestManager_CapturePane(t *testing.T) {
	mock := &mockExecutor{outputs: map[string]string{
		"capture-pane -t my-session -p -S -150": "line 1\nline 2\nline 3",
	}}

	mgr := tmux.NewManager(mock)
	output, err := mgr.CapturePane("my-session", 150)

	assert.NoError(t, err)
	assert.Equal(t, "line 1\nline 2\nline 3", output)
}
```

**Step 10: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run "TestManager_"
```
Expected: FAIL — `tmux.NewManager` 未定義。

**Step 11: 實作 Manager**

在 `internal/tmux/session.go` 追加：
```go
// Manager 管理 tmux session 操作。
type Manager struct {
	exec Executor
}

// NewManager 建立一個新的 tmux Manager。
func NewManager(exec Executor) *Manager {
	return &Manager{exec: exec}
}

// ListSessions 列出所有 tmux sessions。
func (m *Manager) ListSessions() ([]Session, error) {
	output, err := m.exec.Execute("list-sessions", "-F", ListSessionsFormat)
	if err != nil {
		return nil, err
	}
	return ParseListSessions(output)
}

// KillSession 刪除指定的 session。
func (m *Manager) KillSession(name string) error {
	_, err := m.exec.Execute("kill-session", "-t", name)
	return err
}

// RenameSession 重新命名 session。
func (m *Manager) RenameSession(oldName, newName string) error {
	_, err := m.exec.Execute("rename-session", "-t", oldName, newName)
	return err
}

// NewSession 建立新的 detached session。
func (m *Manager) NewSession(name, path string) error {
	_, err := m.exec.Execute("new-session", "-d", "-s", name, "-c", path)
	return err
}

// CapturePane 擷取 session 的終端輸出。
func (m *Manager) CapturePane(name string, lines int) (string, error) {
	return m.exec.Execute("capture-pane", "-t", name, "-p", "-S", fmt.Sprintf("-%d", lines))
}
```

**Step 12: 執行全部測試確認通過**

```bash
go test ./internal/tmux/ -v
```
Expected: 全部 PASS

**Step 13: Commit**

```bash
git add internal/tmux/ go.mod go.sum
git commit -m "feat(tmux): Session 資料模型、時間格式、tmux 操作 Manager"
```

---

## Task 4: 狀態偵測 — 第三層 Terminal Content 解析

最基礎的偵測層，解析終端輸出判斷 Claude Code 狀態。

**Files:**
- Create: `internal/tmux/detector.go`
- Create: `internal/tmux/detector_test.go`

**Step 1: 寫失敗的測試 — ANSI 清除**

Create `internal/tmux/detector_test.go`:
```go
package tmux_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello world", "hello world"},
		{"color codes", "\033[32mgreen\033[0m", "green"},
		{"cursor movement", "\033[2J\033[H", ""},
		{"mixed", "\033[1;33mWarning:\033[0m file not found", "Warning: file not found"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tmux.StripANSI(tt.input))
		})
	}
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestStripANSI
```
Expected: FAIL

**Step 3: 實作 StripANSI（O(n) 單次掃描，避免 regex）**

Create `internal/tmux/detector.go`:
```go
package tmux

import "strings"

// StripANSI 移除 ANSI escape sequences（O(n) 單次掃描）。
func StripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// ESC sequence
			i++
			if i < len(s) && s[i] == '[' {
				// CSI sequence: ESC [ ... (letter)
				i++
				for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
					i++
				}
				if i < len(s) {
					i++ // skip final letter
				}
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/tmux/ -v -run TestStripANSI
```
Expected: PASS

**Step 5: 寫失敗的測試 — DetectStatus 忙碌指標**

在 `internal/tmux/detector_test.go` 追加：
```go
func TestDetectStatus_Running(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"ctrl+c prompt", "Processing files...\n  ctrl+c to interrupt"},
		{"esc prompt", "Reading code...\n  esc to interrupt"},
		{"braille spinner", "⠋ Working on task..."},
		{"asterisk spinner", "* Clauding... (12s, ↓ 200 tokens)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tmux.DetectStatus(tt.content)
			assert.Equal(t, tmux.StatusRunning, status, "content: %q", tt.content)
		})
	}
}

func TestDetectStatus_Waiting(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"prompt char", "Task completed.\n>"},
		{"heavy prompt", "Done.\n❯"},
		{"permission yes/no", "Allow this action?\n  Yes, allow once"},
		{"continue prompt", "Continue? (Y/n)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tmux.DetectStatus(tt.content)
			assert.Equal(t, tmux.StatusWaiting, status, "content: %q", tt.content)
		})
	}
}

func TestDetectStatus_Idle(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"plain shell", "user@host:~$"},
		{"empty", ""},
		{"random text", "some random output\nmore lines"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tmux.DetectStatus(tt.content)
			assert.Equal(t, tmux.StatusIdle, status, "content: %q", tt.content)
		})
	}
}
```

**Step 6: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run "TestDetectStatus_"
```
Expected: FAIL — `tmux.DetectStatus` 未定義。

**Step 7: 實作 DetectStatus**

在 `internal/tmux/detector.go` 追加：
```go
// busyIndicators 表示 AI 正在工作的文字。
var busyIndicators = []string{
	"ctrl+c to interrupt",
	"esc to interrupt",
}

// brailleChars 是 Claude Code spinner 使用的 Braille 字元範圍。
var brailleChars = [...]rune{
	'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏',
}

// waitingIndicators 表示需要人類輸入的文字。
var waitingIndicators = []string{
	"Yes, allow once",
	"No, and tell Claude",
	"Continue? (Y/n)",
	"(Y/n)",
}

// DetectStatus 從終端內容偵測 Claude Code 的狀態。
func DetectStatus(content string) SessionStatus {
	if content == "" {
		return StatusIdle
	}

	clean := StripANSI(content)

	// 檢查忙碌指標
	for _, indicator := range busyIndicators {
		if strings.Contains(clean, indicator) {
			return StatusRunning
		}
	}

	// 檢查 braille spinner
	for _, r := range clean {
		for _, br := range brailleChars {
			if r == br {
				return StatusRunning
			}
		}
	}

	// 檢查 asterisk spinner + whimsical words
	if containsSpinnerPattern(clean) {
		return StatusRunning
	}

	// 檢查等待指標
	for _, indicator := range waitingIndicators {
		if strings.Contains(clean, indicator) {
			return StatusWaiting
		}
	}

	// 檢查最後一行的 prompt 字元
	if hasPromptChar(clean) {
		return StatusWaiting
	}

	return StatusIdle
}

// containsSpinnerPattern 檢查 asterisk spinner 模式（如 "* Clauding..."）。
func containsSpinnerPattern(s string) bool {
	return strings.Contains(s, "* Clauding") ||
		strings.Contains(s, "* Hullaballooing") ||
		strings.Contains(s, "* Thinking") ||
		strings.Contains(s, "* Pondering")
}

// hasPromptChar 檢查最後一行是否為 prompt 字元。
func hasPromptChar(s string) bool {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) == 0 {
		return false
	}
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	return lastLine == ">" || lastLine == "❯"
}
```

**Step 8: 執行全部測試確認通過**

```bash
go test ./internal/tmux/ -v
```
Expected: 全部 PASS

**Step 9: Commit**

```bash
git add internal/tmux/detector.go internal/tmux/detector_test.go
git commit -m "feat(tmux): 終端內容狀態偵測（第三層）與 ANSI 清除"
```

---

## Task 5: 狀態偵測 — 第二層 Pane Title

從 tmux pane title 偵測 Claude Code 狀態。

**Files:**
- Modify: `internal/tmux/detector.go`
- Modify: `internal/tmux/detector_test.go`

**Step 1: 寫失敗的測試**

在 `internal/tmux/detector_test.go` 追加：
```go
func TestDetectTitleStatus(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected tmux.TitleStatus
	}{
		{"braille spinner", "⠋ Claude Code", tmux.TitleRunning},
		{"another braille", "⠹ Processing", tmux.TitleRunning},
		{"checkmark", "✓ Done", tmux.TitleDone},
		{"check emoji", "✔ Complete", tmux.TitleDone},
		{"no indicator", "my-session", tmux.TitleUnknown},
		{"empty", "", tmux.TitleUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tmux.DetectTitleStatus(tt.title))
		})
	}
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestDetectTitleStatus
```
Expected: FAIL

**Step 3: 實作 DetectTitleStatus**

在 `internal/tmux/detector.go` 追加：
```go
// TitleStatus 是 pane title 偵測的結果。
type TitleStatus int

const (
	TitleUnknown TitleStatus = iota
	TitleRunning             // 偵測到 spinner
	TitleDone                // 偵測到完成標記
)

// DetectTitleStatus 從 pane title 偵測狀態。
func DetectTitleStatus(title string) TitleStatus {
	for _, r := range title {
		// Braille spinner range: U+2800-U+28FF
		if r >= 0x2800 && r <= 0x28FF {
			return TitleRunning
		}
		// Checkmark characters
		if r == '✓' || r == '✔' {
			return TitleDone
		}
	}
	return TitleUnknown
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/tmux/ -v -run TestDetectTitleStatus
```
Expected: PASS

**Step 5: 寫失敗的測試 — ParseListPaneTitles**

在 `internal/tmux/detector_test.go` 追加：
```go
func TestParseListPaneTitles(t *testing.T) {
	output := `my-project:⠋ Working
api-server:✓ Done
frontend:bash`

	titles, err := tmux.ParseListPaneTitles(output)
	assert.NoError(t, err)
	assert.Len(t, titles, 3)
	assert.Equal(t, "⠋ Working", titles["my-project"])
	assert.Equal(t, "✓ Done", titles["api-server"])
	assert.Equal(t, "bash", titles["frontend"])
}
```

**Step 6: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestParseListPaneTitles
```
Expected: FAIL

**Step 7: 實作 ParseListPaneTitles**

在 `internal/tmux/detector.go` 追加：
```go
import "fmt"

// PaneTitleFormat 是 tmux list-panes -a -F 的格式字串。
const PaneTitleFormat = "#{session_name}:#{pane_title}"

// ParseListPaneTitles 解析 list-panes 輸出為 session→title 的 map。
func ParseListPaneTitles(output string) (map[string]string, error) {
	result := make(map[string]string)
	output = strings.TrimSpace(output)
	if output == "" {
		return result, nil
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			return nil, fmt.Errorf("unexpected format: %q", line)
		}
		name := line[:idx]
		title := line[idx+1:]
		result[name] = title
	}
	return result, nil
}
```

**Step 8: 執行全部測試確認通過**

```bash
go test ./internal/tmux/ -v
```
Expected: 全部 PASS

**Step 9: Commit**

```bash
git add internal/tmux/
git commit -m "feat(tmux): Pane title 狀態偵測（第二層）"
```

---

## Task 6: 狀態偵測 — 第一層 Hook 檔案監聽

讀取 Claude Code hook 寫入的狀態檔案。

**Files:**
- Create: `internal/tmux/hookstatus.go`
- Create: `internal/tmux/hookstatus_test.go`

**Step 1: 寫失敗的測試 — 讀取狀態檔案**

Create `internal/tmux/hookstatus_test.go`:
```go
package tmux_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestReadHookStatus(t *testing.T) {
	dir := t.TempDir()

	// 建立狀態檔案
	statusFile := filepath.Join(dir, "my-session")
	content := `{"status":"running","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `,"event":"UserPromptSubmit"}`
	require.NoError(t, os.WriteFile(statusFile, []byte(content), 0644))

	hs, err := tmux.ReadHookStatus(dir, "my-session")
	require.NoError(t, err)

	assert.Equal(t, tmux.StatusRunning, hs.Status)
	assert.Equal(t, "UserPromptSubmit", hs.Event)
	assert.True(t, hs.IsValid()) // 2 分鐘內有效
}

func TestReadHookStatus_Expired(t *testing.T) {
	dir := t.TempDir()

	// 建立過期的狀態檔案（3 分鐘前）
	statusFile := filepath.Join(dir, "old-session")
	oldTime := time.Now().Add(-3 * time.Minute).Unix()
	content := fmt.Sprintf(`{"status":"running","timestamp":%d,"event":"UserPromptSubmit"}`, oldTime)
	require.NoError(t, os.WriteFile(statusFile, []byte(content), 0644))

	hs, err := tmux.ReadHookStatus(dir, "old-session")
	require.NoError(t, err)

	assert.False(t, hs.IsValid()) // 超過 2 分鐘，無效
}

func TestReadHookStatus_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := tmux.ReadHookStatus(dir, "nonexistent")
	assert.Error(t, err)
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestReadHookStatus
```
Expected: FAIL

**Step 3: 實作 ReadHookStatus**

Create `internal/tmux/hookstatus.go`:
```go
package tmux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const hookStatusTTL = 2 * time.Minute

// HookStatus 代表 Claude Code hook 寫入的狀態。
type HookStatus struct {
	Status    SessionStatus `json:"-"`
	RawStatus string        `json:"status"`
	Timestamp int64         `json:"timestamp"`
	Event     string        `json:"event"`
}

// IsValid 檢查狀態是否在 TTL 內有效。
func (h HookStatus) IsValid() bool {
	return time.Since(time.Unix(h.Timestamp, 0)) < hookStatusTTL
}

// ReadHookStatus 讀取指定 session 的 hook 狀態檔案。
func ReadHookStatus(statusDir, sessionName string) (HookStatus, error) {
	path := filepath.Join(statusDir, sessionName)
	data, err := os.ReadFile(path)
	if err != nil {
		return HookStatus{}, fmt.Errorf("read hook status: %w", err)
	}

	var hs HookStatus
	if err := json.Unmarshal(data, &hs); err != nil {
		return HookStatus{}, fmt.Errorf("parse hook status: %w", err)
	}

	// 將字串狀態轉換為 SessionStatus
	switch hs.RawStatus {
	case "running":
		hs.Status = StatusRunning
	case "waiting":
		hs.Status = StatusWaiting
	case "error":
		hs.Status = StatusError
	default:
		hs.Status = StatusIdle
	}

	return hs, nil
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/tmux/ -v -run TestReadHookStatus
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tmux/hookstatus.go internal/tmux/hookstatus_test.go
git commit -m "feat(tmux): Hook 狀態檔案讀取（第一層偵測）"
```

---

## Task 7: 狀態偵測 — 三層整合 StatusResolver

將三層偵測合併為統一的狀態解析器。

**Files:**
- Create: `internal/tmux/resolver.go`
- Create: `internal/tmux/resolver_test.go`

**Step 1: 寫失敗的測試**

Create `internal/tmux/resolver_test.go`:
```go
package tmux_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestResolveStatus_HookTakesPriority(t *testing.T) {
	input := tmux.StatusInput{
		HookStatus:   &tmux.HookStatus{Status: tmux.StatusRunning, Timestamp: timeNowUnix(), RawStatus: "running"},
		PaneTitle:    "",
		PaneContent:  ">", // 會被判定為 waiting，但 hook 優先
	}

	status := tmux.ResolveStatus(input)
	assert.Equal(t, tmux.StatusRunning, status)
}

func TestResolveStatus_TitleFallback(t *testing.T) {
	input := tmux.StatusInput{
		HookStatus:  nil,           // 無 hook
		PaneTitle:   "⠋ Working",  // title 顯示 running
		PaneContent: "",
	}

	status := tmux.ResolveStatus(input)
	assert.Equal(t, tmux.StatusRunning, status)
}

func TestResolveStatus_TitleDone_FallsToContent(t *testing.T) {
	input := tmux.StatusInput{
		HookStatus:  nil,
		PaneTitle:   "✓ Done",
		PaneContent: ">", // prompt 表示 waiting
	}

	status := tmux.ResolveStatus(input)
	assert.Equal(t, tmux.StatusWaiting, status)
}

func TestResolveStatus_ContentFallback(t *testing.T) {
	input := tmux.StatusInput{
		HookStatus:  nil,
		PaneTitle:   "bash",                          // 無指標
		PaneContent: "Processing\n  ctrl+c to interrupt", // running
	}

	status := tmux.ResolveStatus(input)
	assert.Equal(t, tmux.StatusRunning, status)
}

func TestResolveStatus_AllEmpty(t *testing.T) {
	input := tmux.StatusInput{
		HookStatus:  nil,
		PaneTitle:   "",
		PaneContent: "user@host:~$",
	}

	status := tmux.ResolveStatus(input)
	assert.Equal(t, tmux.StatusIdle, status)
}

func TestResolveStatus_ExpiredHook_Ignored(t *testing.T) {
	// Hook 過期時應該被忽略，fallback 到下一層
	expiredHook := &tmux.HookStatus{
		Status:    tmux.StatusRunning,
		Timestamp: timeNowUnix() - 180, // 3 分鐘前
		RawStatus: "running",
	}

	input := tmux.StatusInput{
		HookStatus:  expiredHook,
		PaneTitle:   "",
		PaneContent: ">",
	}

	status := tmux.ResolveStatus(input)
	assert.Equal(t, tmux.StatusWaiting, status) // 使用 content 層
}

func timeNowUnix() int64 {
	return time.Now().Unix()
}
```

注意：需要在檔案頂部加 `import "time"`。

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestResolveStatus
```
Expected: FAIL

**Step 3: 實作 ResolveStatus**

Create `internal/tmux/resolver.go`:
```go
package tmux

// StatusInput 包含三層偵測的輸入資料。
type StatusInput struct {
	HookStatus  *HookStatus // 第一層：hook 狀態（可能為 nil）
	PaneTitle   string      // 第二層：pane title
	PaneContent string      // 第三層：終端內容
}

// ResolveStatus 依優先順序從三層偵測中解析最終狀態。
func ResolveStatus(input StatusInput) SessionStatus {
	// 第一層：Hook（最高優先，2 分鐘內有效）
	if input.HookStatus != nil && input.HookStatus.IsValid() {
		return input.HookStatus.Status
	}

	// 第二層：Pane title
	titleStatus := DetectTitleStatus(input.PaneTitle)
	switch titleStatus {
	case TitleRunning:
		return StatusRunning
	case TitleDone:
		// Title 顯示完成，需要用第三層確認是否 waiting
		return DetectStatus(input.PaneContent)
	}

	// 第三層：Terminal content（fallback）
	return DetectStatus(input.PaneContent)
}
```

**Step 4: 執行全部測試確認通過**

```bash
go test ./internal/tmux/ -v
```
Expected: 全部 PASS

**Step 5: Commit**

```bash
git add internal/tmux/resolver.go internal/tmux/resolver_test.go
git commit -m "feat(tmux): 三層狀態偵測整合 StatusResolver"
```

---

## Task 8: Store 模組 — SQLite 群組與排序

SQLite 儲存群組和排序資訊。

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/store_test.go`

**Step 1: 安裝 SQLite 依賴**

```bash
go get modernc.org/sqlite
```

**Step 2: 寫失敗的測試 — 群組 CRUD**

Create `internal/store/store_test.go`:
```go
package store_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wake/tmux-session-menu/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestGroup_CRUD(t *testing.T) {
	s := newTestStore(t)

	// Create
	err := s.CreateGroup("工作專案", 0)
	require.NoError(t, err)

	err = s.CreateGroup("維運", 1)
	require.NoError(t, err)

	// Read
	groups, err := s.ListGroups()
	require.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.Equal(t, "工作專案", groups[0].Name)
	assert.Equal(t, 0, groups[0].SortOrder)
	assert.Equal(t, "維運", groups[1].Name)

	// Update — 更名
	err = s.RenameGroup(groups[0].ID, "開發專案")
	require.NoError(t, err)

	groups, err = s.ListGroups()
	require.NoError(t, err)
	assert.Equal(t, "開發專案", groups[0].Name)

	// Delete
	err = s.DeleteGroup(groups[1].ID)
	require.NoError(t, err)

	groups, err = s.ListGroups()
	require.NoError(t, err)
	assert.Len(t, groups, 1)
}
```

**Step 3: 執行測試確認失敗**

```bash
go test ./internal/store/ -v -run TestGroup_CRUD
```
Expected: FAIL

**Step 4: 實作 Store 與群組功能**

Create `internal/store/store.go`:
```go
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Group 代表一個 session 群組。
type Group struct {
	ID        int64
	Name      string
	SortOrder int
	Collapsed bool
}

// SessionMeta 儲存 session 的額外資訊（群組歸屬、排序等）。
type SessionMeta struct {
	SessionName string
	GroupID     int64 // 0 = 未分組
	SortOrder   int
	CustomName  string // 使用者自訂的顯示名稱
}

// Store 管理 SQLite 資料存取。
type Store struct {
	db *sql.DB
}

// Open 開啟或建立 SQLite 資料庫。
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// Close 關閉資料庫。
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		sort_order INTEGER NOT NULL DEFAULT 0,
		collapsed INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS session_meta (
		session_name TEXT PRIMARY KEY,
		group_id INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		custom_name TEXT NOT NULL DEFAULT ''
	);`
	_, err := s.db.Exec(schema)
	return err
}

// CreateGroup 建立新群組。
func (s *Store) CreateGroup(name string, sortOrder int) error {
	_, err := s.db.Exec("INSERT INTO groups (name, sort_order) VALUES (?, ?)", name, sortOrder)
	return err
}

// ListGroups 列出所有群組，依排序順序。
func (s *Store) ListGroups() ([]Group, error) {
	rows, err := s.db.Query("SELECT id, name, sort_order, collapsed FROM groups ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.SortOrder, &g.Collapsed); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// RenameGroup 重新命名群組。
func (s *Store) RenameGroup(id int64, name string) error {
	_, err := s.db.Exec("UPDATE groups SET name = ? WHERE id = ?", name, id)
	return err
}

// DeleteGroup 刪除群組。
func (s *Store) DeleteGroup(id int64) error {
	// 將屬於該群組的 session 移到未分組
	_, _ = s.db.Exec("UPDATE session_meta SET group_id = 0 WHERE group_id = ?", id)
	_, err := s.db.Exec("DELETE FROM groups WHERE id = ?", id)
	return err
}
```

**Step 5: 執行測試確認通過**

```bash
go test ./internal/store/ -v -run TestGroup_CRUD
```
Expected: PASS

**Step 6: 寫失敗的測試 — Session 群組歸屬與排序**

在 `internal/store/store_test.go` 追加：
```go
func TestSessionMeta_AssignAndList(t *testing.T) {
	s := newTestStore(t)

	// 建立群組
	require.NoError(t, s.CreateGroup("dev", 0))
	groups, _ := s.ListGroups()
	groupID := groups[0].ID

	// 將 session 歸入群組
	require.NoError(t, s.SetSessionGroup("my-project", groupID, 0))
	require.NoError(t, s.SetSessionGroup("api-server", groupID, 1))
	require.NoError(t, s.SetSessionGroup("standalone", 0, 0)) // 未分組

	// 列出群組內的 sessions
	metas, err := s.ListSessionMetas(groupID)
	require.NoError(t, err)
	assert.Len(t, metas, 2)
	assert.Equal(t, "my-project", metas[0].SessionName)
	assert.Equal(t, "api-server", metas[1].SessionName)

	// 列出未分組的 sessions
	ungrouped, err := s.ListSessionMetas(0)
	require.NoError(t, err)
	assert.Len(t, ungrouped, 1)
	assert.Equal(t, "standalone", ungrouped[0].SessionName)
}

func TestSessionMeta_Reorder(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.CreateGroup("dev", 0))
	groups, _ := s.ListGroups()
	gid := groups[0].ID

	require.NoError(t, s.SetSessionGroup("a", gid, 0))
	require.NoError(t, s.SetSessionGroup("b", gid, 1))
	require.NoError(t, s.SetSessionGroup("c", gid, 2))

	// 將 c 移到最前面
	require.NoError(t, s.SetSessionGroup("c", gid, -1))
	require.NoError(t, s.SetSessionGroup("a", gid, 0))
	require.NoError(t, s.SetSessionGroup("b", gid, 1))

	metas, _ := s.ListSessionMetas(gid)
	assert.Equal(t, "c", metas[0].SessionName)
	assert.Equal(t, "a", metas[1].SessionName)
	assert.Equal(t, "b", metas[2].SessionName)
}

func TestGroup_Reorder(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.CreateGroup("first", 0))
	require.NoError(t, s.CreateGroup("second", 1))
	require.NoError(t, s.CreateGroup("third", 2))

	groups, _ := s.ListGroups()

	// 將 third 移到最前面
	require.NoError(t, s.SetGroupOrder(groups[2].ID, -1))
	require.NoError(t, s.SetGroupOrder(groups[0].ID, 0))
	require.NoError(t, s.SetGroupOrder(groups[1].ID, 1))

	groups, _ = s.ListGroups()
	assert.Equal(t, "third", groups[0].Name)
	assert.Equal(t, "first", groups[1].Name)
	assert.Equal(t, "second", groups[2].Name)
}
```

**Step 7: 執行測試確認失敗**

```bash
go test ./internal/store/ -v -run "TestSessionMeta_|TestGroup_Reorder"
```
Expected: FAIL

**Step 8: 實作 Session meta 與排序**

在 `internal/store/store.go` 追加：
```go
// SetSessionGroup 設定 session 的群組和排序。
func (s *Store) SetSessionGroup(sessionName string, groupID int64, sortOrder int) error {
	_, err := s.db.Exec(`
		INSERT INTO session_meta (session_name, group_id, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT(session_name) DO UPDATE SET group_id = ?, sort_order = ?`,
		sessionName, groupID, sortOrder, groupID, sortOrder)
	return err
}

// ListSessionMetas 列出指定群組的 session meta，依排序順序。
func (s *Store) ListSessionMetas(groupID int64) ([]SessionMeta, error) {
	rows, err := s.db.Query(
		"SELECT session_name, group_id, sort_order, custom_name FROM session_meta WHERE group_id = ? ORDER BY sort_order, session_name",
		groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metas []SessionMeta
	for rows.Next() {
		var m SessionMeta
		if err := rows.Scan(&m.SessionName, &m.GroupID, &m.SortOrder, &m.CustomName); err != nil {
			return nil, err
		}
		metas = append(metas, m)
	}
	return metas, rows.Err()
}

// SetGroupOrder 設定群組的排序順序。
func (s *Store) SetGroupOrder(id int64, sortOrder int) error {
	_, err := s.db.Exec("UPDATE groups SET sort_order = ? WHERE id = ?", sortOrder, id)
	return err
}
```

**Step 9: 執行全部測試確認通過**

```bash
go test ./internal/store/ -v
```
Expected: 全部 PASS

**Step 10: Commit**

```bash
git add internal/store/ go.mod go.sum
git commit -m "feat(store): SQLite 群組和 session 排序儲存"
```

---

## Task 9: Bubble Tea TUI — 基礎 App 架構

建立 Bubble Tea 應用的骨架：Model、Update、View。

**Files:**
- Create: `internal/ui/app.go`
- Create: `internal/ui/app_test.go`
- Create: `internal/ui/styles.go`

**Step 1: 安裝 Bubble Tea 依賴**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

**Step 2: 寫失敗的測試 — Model 初始化**

Create `internal/ui/app_test.go`:
```go
package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/ui"
)

func TestModel_Init(t *testing.T) {
	m := ui.NewModel()

	// Init 應回傳一個 batch of commands
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestModel_View_ShowsHeader(t *testing.T) {
	m := ui.NewModel()

	view := m.View()
	assert.Contains(t, view, "tmux session menu")
}

func TestModel_Quit(t *testing.T) {
	m := ui.NewModel()

	// 按 q 應該觸發退出
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated

	// cmd 應該包含 Quit
	assert.NotNil(t, cmd)
}
```

**Step 3: 執行測試確認失敗**

```bash
go test ./internal/ui/ -v -run "TestModel_"
```
Expected: FAIL

**Step 4: 實作基礎 Model**

Create `internal/ui/styles.go`:
```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#c0caf5"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7aa2f7")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#787fa0"))

	statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	statusWaiting = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68"))
	statusIdle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#787fa0"))
	statusError   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))

	previewBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("#414868"))
)
```

Create `internal/ui/app.go`:
```go
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Model 是 Bubble Tea 的主 model。
type Model struct {
	width  int
	height int
	cursor int
	quitting bool
}

// NewModel 建立新的 Model。
func NewModel() Model {
	return Model{}
}

// Init 實作 tea.Model。
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("tmux session menu")
}

// Update 實作 tea.Model。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View 實作 tea.Model。
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	header := headerStyle.Render("tmux session menu") +
		dimStyle.Render("  (↑↓/jk 選擇, Enter 確認, q 離開)")

	help := fmt.Sprintf("\n  %s  %s  %s",
		dimStyle.Render("[n] 新建"),
		dimStyle.Render("[g] 新群組"),
		dimStyle.Render("[q] 離開"),
	)

	return header + "\n" + help + "\n"
}
```

**Step 5: 執行測試確認通過**

```bash
go test ./internal/ui/ -v -run "TestModel_"
```
Expected: PASS

**Step 6: Commit**

```bash
git add internal/ui/ go.mod go.sum
git commit -m "feat(ui): Bubble Tea 基礎 Model 骨架與樣式"
```

---

## Task 10: TUI — Session 列表顯示與導航

將真實 session 資料整合到 TUI 列表。

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/app_test.go`
- Create: `internal/ui/items.go`
- Create: `internal/ui/items_test.go`

**Step 1: 寫失敗的測試 — 扁平化群組+session 為可導航列表**

Create `internal/ui/items_test.go`:
```go
package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/store"
	"github.com/wake/tmux-session-menu/internal/tmux"
	"github.com/wake/tmux-session-menu/internal/ui"
)

func TestFlattenItems(t *testing.T) {
	groups := []store.Group{
		{ID: 1, Name: "dev", SortOrder: 0},
		{ID: 2, Name: "ops", SortOrder: 1},
	}

	sessions := []tmux.Session{
		{Name: "project-a", GroupName: "dev"},
		{Name: "project-b", GroupName: "dev"},
		{Name: "monitoring", GroupName: "ops"},
		{Name: "standalone"}, // 未分組
	}

	items := ui.FlattenItems(groups, sessions)

	// 未分組的 session 在最上面
	assert.Equal(t, ui.ItemSession, items[0].Type)
	assert.Equal(t, "standalone", items[0].Session.Name)

	// 然後是群組
	assert.Equal(t, ui.ItemGroup, items[1].Type)
	assert.Equal(t, "dev", items[1].Group.Name)

	// 群組下的 sessions
	assert.Equal(t, ui.ItemSession, items[2].Type)
	assert.Equal(t, "project-a", items[2].Session.Name)
	assert.Equal(t, ui.ItemSession, items[3].Type)
	assert.Equal(t, "project-b", items[3].Session.Name)

	// 第二個群組
	assert.Equal(t, ui.ItemGroup, items[4].Type)
	assert.Equal(t, "ops", items[4].Group.Name)
	assert.Equal(t, ui.ItemSession, items[5].Type)
	assert.Equal(t, "monitoring", items[5].Session.Name)
}

func TestFlattenItems_CollapsedGroup(t *testing.T) {
	groups := []store.Group{
		{ID: 1, Name: "dev", SortOrder: 0, Collapsed: true},
	}

	sessions := []tmux.Session{
		{Name: "project-a", GroupName: "dev"},
		{Name: "project-b", GroupName: "dev"},
	}

	items := ui.FlattenItems(groups, sessions)

	// 折疊群組只顯示群組本身，不顯示子項目
	assert.Len(t, items, 1)
	assert.Equal(t, ui.ItemGroup, items[0].Type)
	assert.Equal(t, "dev", items[0].Group.Name)
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/ui/ -v -run TestFlattenItems
```
Expected: FAIL

**Step 3: 實作 FlattenItems**

Create `internal/ui/items.go`:
```go
package ui

import (
	"github.com/wake/tmux-session-menu/internal/store"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

// ItemType 表示列表項目的類型。
type ItemType int

const (
	ItemSession ItemType = iota
	ItemGroup
)

// ListItem 是列表中可導航的項目。
type ListItem struct {
	Type    ItemType
	Session tmux.Session // 當 Type == ItemSession 時有效
	Group   store.Group  // 當 Type == ItemGroup 時有效
}

// FlattenItems 將群組和 sessions 合併為扁平的可導航列表。
// 順序：未分組 sessions → 群組1 → 群組1 的 sessions → 群組2 → ...
func FlattenItems(groups []store.Group, sessions []tmux.Session) []ListItem {
	var items []ListItem

	// 建立 group name → sessions 的 map
	grouped := make(map[string][]tmux.Session)
	var ungrouped []tmux.Session

	for _, s := range sessions {
		if s.GroupName == "" {
			ungrouped = append(ungrouped, s)
		} else {
			grouped[s.GroupName] = append(grouped[s.GroupName], s)
		}
	}

	// 未分組的 sessions
	for _, s := range ungrouped {
		items = append(items, ListItem{Type: ItemSession, Session: s})
	}

	// 按群組排序加入
	for _, g := range groups {
		items = append(items, ListItem{Type: ItemGroup, Group: g})

		// 如果群組未折疊，加入群組下的 sessions
		if !g.Collapsed {
			for _, s := range grouped[g.Name] {
				items = append(items, ListItem{Type: ItemSession, Session: s})
			}
		}
	}

	return items
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/ui/ -v -run TestFlattenItems
```
Expected: PASS

**Step 5: 寫失敗的測試 — 鍵盤導航**

在 `internal/ui/app_test.go` 追加：
```go
func TestModel_Navigation(t *testing.T) {
	m := ui.NewModel()
	m.SetItems([]ui.ListItem{
		{Type: ui.ItemSession},
		{Type: ui.ItemSession},
		{Type: ui.ItemSession},
	})

	// 初始 cursor 在 0
	assert.Equal(t, 0, m.Cursor())

	// j 往下
	m, _ = applyKey(m, "j")
	assert.Equal(t, 1, m.Cursor())

	// k 往上
	m, _ = applyKey(m, "k")
	assert.Equal(t, 0, m.Cursor())

	// 不能超過邊界
	m, _ = applyKey(m, "k")
	assert.Equal(t, 0, m.Cursor())

	// 到底部
	m, _ = applyKey(m, "j")
	m, _ = applyKey(m, "j")
	m, _ = applyKey(m, "j") // 超過邊界
	assert.Equal(t, 2, m.Cursor())
}

func applyKey(m ui.Model, key string) (ui.Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	updated, cmd := m.Update(msg)
	return updated.(ui.Model), cmd
}
```

**Step 6: 執行測試確認失敗**

```bash
go test ./internal/ui/ -v -run TestModel_Navigation
```
Expected: FAIL — `SetItems` 和 `Cursor` 未定義。

**Step 7: 實作導航功能**

在 `internal/ui/app.go` 中修改 Model：

將 Model 結構和方法更新為：
```go
// Model 是 Bubble Tea 的主 model。
type Model struct {
	width    int
	height   int
	cursor   int
	items    []ListItem
	quitting bool
}

// SetItems 設定列表項目（主要用於測試）。
func (m *Model) SetItems(items []ListItem) {
	m.items = items
	if m.cursor >= len(items) && len(items) > 0 {
		m.cursor = len(items) - 1
	}
}

// Cursor 回傳目前的游標位置。
func (m Model) Cursor() int {
	return m.cursor
}
```

在 Update 方法的 KeyMsg switch 中加入：
```go
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
```

**Step 8: 執行全部測試確認通過**

```bash
go test ./internal/ui/ -v
```
Expected: 全部 PASS

**Step 9: Commit**

```bash
git add internal/ui/
git commit -m "feat(ui): Session 列表扁平化與鍵盤導航"
```

---

## Task 11: TUI — 列表渲染與預覽

將 session 列表和預覽區渲染到終端。

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/app_test.go`

**Step 1: 寫失敗的測試 — View 渲染包含 session 資訊**

在 `internal/ui/app_test.go` 追加：
```go
import "time"

func TestModel_View_RendersSessions(t *testing.T) {
	m := ui.NewModel()
	m.SetItems([]ui.ListItem{
		{Type: ui.ItemGroup, Group: store.Group{Name: "dev"}},
		{Type: ui.ItemSession, Session: tmux.Session{
			Name:     "my-project",
			Status:   tmux.StatusRunning,
			Activity: time.Now().Add(-3 * time.Minute),
			AIModel:  "claude-sonnet-4-6",
		}},
		{Type: ui.ItemSession, Session: tmux.Session{
			Name:     "api-server",
			Status:   tmux.StatusIdle,
			Activity: time.Now().Add(-2 * time.Hour),
		}},
	})

	view := m.View()

	// 應包含群組名
	assert.Contains(t, view, "dev")
	// 應包含 session 名
	assert.Contains(t, view, "my-project")
	assert.Contains(t, view, "api-server")
	// 應包含狀態圖示
	assert.Contains(t, view, "●") // running
	assert.Contains(t, view, "○") // idle
	// 應包含 AI 模型
	assert.Contains(t, view, "claude-sonnet-4-6")
}

func TestModel_View_Preview(t *testing.T) {
	m := ui.NewModel()
	m.SetItems([]ui.ListItem{
		{Type: ui.ItemSession, Session: tmux.Session{
			Name:      "my-project",
			AISummary: "正在重構 auth 模組",
		}},
	})

	view := m.View()
	assert.Contains(t, view, "正在重構 auth 模組")
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/ui/ -v -run "TestModel_View_Renders|TestModel_View_Preview"
```
Expected: FAIL

**Step 3: 實作 View 渲染**

更新 `internal/ui/app.go` 的 View 方法：

```go
import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("tmux session menu"))
	b.WriteString(dimStyle.Render("  (↑↓/jk 選擇, Enter 確認, q 離開)"))
	b.WriteString("\n\n")

	// Items
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "► "
		}

		switch item.Type {
		case ItemGroup:
			arrow := "▼"
			if item.Group.Collapsed {
				arrow = "▶"
			}
			if i == m.cursor {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("%s%s %s", cursor, arrow, item.Group.Name)))
			} else {
				b.WriteString(dimStyle.Render(fmt.Sprintf("%s%s %s", cursor, arrow, item.Group.Name)))
			}
			b.WriteString("\n")

		case ItemSession:
			s := item.Session
			statusStyle := statusStyleFor(s.Status)
			icon := statusStyle.Render(s.StatusIcon())
			relTime := dimStyle.Render(s.RelativeTime())

			line := fmt.Sprintf("%s  %s %s  %s", cursor, s.Name, icon, relTime)
			if s.AIModel != "" {
				line += "  " + dimStyle.Render(s.AIModel)
			}

			if i == m.cursor {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("%s  %s", cursor, s.Name)))
				b.WriteString(fmt.Sprintf(" %s  %s", icon, relTime))
				if s.AIModel != "" {
					b.WriteString("  " + dimStyle.Render(s.AIModel))
				}
			} else {
				b.WriteString(fmt.Sprintf("%s  %s %s  %s", cursor, s.Name, icon, relTime))
				if s.AIModel != "" {
					b.WriteString("  " + dimStyle.Render(s.AIModel))
				}
			}
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString(fmt.Sprintf("\n  %s  %s  %s\n",
		dimStyle.Render("[n] 新建"),
		dimStyle.Render("[g] 新群組"),
		dimStyle.Render("[q] 離開"),
	))

	// Preview
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := m.items[m.cursor]
		if item.Type == ItemSession {
			preview := item.Session.AISummary
			if preview != "" {
				b.WriteString("\n")
				b.WriteString(previewBorder.Render(fmt.Sprintf("Preview: %s", preview)))
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func statusStyleFor(status tmux.SessionStatus) lipgloss.Style {
	switch status {
	case tmux.StatusRunning:
		return statusRunning
	case tmux.StatusWaiting:
		return statusWaiting
	case tmux.StatusError:
		return statusError
	default:
		return statusIdle
	}
}
```

**Step 4: 執行全部測試確認通過**

```bash
go test ./internal/ui/ -v
```
Expected: 全部 PASS

**Step 5: Commit**

```bash
git add internal/ui/
git commit -m "feat(ui): Session 列表渲染與底部預覽"
```

---

## Task 12: CLI 入口 — 整合全部模組

將所有模組串接到 cmd/tsm/main.go。

**Files:**
- Modify: `cmd/tsm/main.go`
- Create: `internal/tmux/executor.go`

**Step 1: 寫失敗的測試 — 真實 Executor**

Create `internal/tmux/executor_test.go`:
```go
package tmux_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestRealExecutor_TmuxNotRunning(t *testing.T) {
	// 即使 tmux 沒跑，Execute 也不應 panic
	exec := tmux.NewRealExecutor()
	_, err := exec.Execute("list-sessions")
	// 允許失敗（tmux 可能沒在跑），但不應 panic
	_ = err
}

func TestTmuxAvailable(t *testing.T) {
	_, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux not installed")
	}
	assert.NoError(t, err)
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/tmux/ -v -run TestRealExecutor
```
Expected: FAIL — `tmux.NewRealExecutor` 未定義。

**Step 3: 實作 RealExecutor**

Create `internal/tmux/executor.go`:
```go
package tmux

import (
	"os/exec"
	"strings"
)

// RealExecutor 使用真實的 tmux 指令。
type RealExecutor struct{}

// NewRealExecutor 建立 RealExecutor。
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Execute 執行 tmux 指令。
func (e *RealExecutor) Execute(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/tmux/ -v -run TestRealExecutor
```
Expected: PASS

**Step 5: 整合 main.go**

Update `cmd/tsm/main.go`:
```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wake/tmux-session-menu/internal/ui"
)

func main() {
	m := ui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 6: 驗證編譯**

```bash
make build
```
Expected: 成功產生 `tsm` 二進位。

**Step 7: 執行全部測試**

```bash
make test
```
Expected: 全部 PASS

**Step 8: Commit**

```bash
git add cmd/tsm/main.go internal/tmux/executor.go internal/tmux/executor_test.go
git commit -m "feat: 整合 CLI 入口，串接所有模組"
```

---

## Task 13: AI 模型偵測

從終端內容偵測 Claude Code 使用的 AI 模型。

**Files:**
- Create: `internal/ai/status.go`
- Create: `internal/ai/status_test.go`

**Step 1: 寫失敗的測試**

Create `internal/ai/status_test.go`:
```go
package ai_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/ai"
)

func TestDetectModel(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"sonnet model", "Using claude-sonnet-4-6\n> ", "claude-sonnet-4-6"},
		{"opus model", "Model: claude-opus-4-6\nProcessing...", "claude-opus-4-6"},
		{"haiku model", "claude-haiku-4-5-20251001 ready", "claude-haiku-4-5-20251001"},
		{"no model", "regular shell output", ""},
		{"status line model", "╭ claude-sonnet-4-6 · $0.02", "claude-sonnet-4-6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ai.DetectModel(tt.content))
		})
	}
}

func TestDetectTool(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"claude code prompt", "some output\n>", "claude-code"},
		{"claude code interrupt", "ctrl+c to interrupt", "claude-code"},
		{"plain shell", "user@host:~$", ""},
		{"gemini", "gemini>", ""},  // 暫不支援
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ai.DetectTool(tt.content))
		})
	}
}
```

**Step 2: 執行測試確認失敗**

```bash
go test ./internal/ai/ -v
```
Expected: FAIL

**Step 3: 實作偵測**

Create `internal/ai/status.go`:
```go
package ai

import (
	"regexp"
	"strings"
)

// Claude model 的正則表達式。
var claudeModelPattern = regexp.MustCompile(`claude-(?:sonnet|opus|haiku)-[\w.-]+`)

// DetectModel 從終端內容偵測 AI 模型名稱。
func DetectModel(content string) string {
	match := claudeModelPattern.FindString(content)
	return match
}

// DetectTool 從終端內容偵測使用的 AI 工具。
func DetectTool(content string) string {
	// Claude Code 特徵
	claudeIndicators := []string{
		"ctrl+c to interrupt",
		"esc to interrupt",
	}

	for _, indicator := range claudeIndicators {
		if strings.Contains(content, indicator) {
			return "claude-code"
		}
	}

	// 檢查 Claude Code prompt
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) > 0 {
		last := strings.TrimSpace(lines[len(lines)-1])
		if last == ">" || last == "❯" {
			return "claude-code"
		}
	}

	return ""
}
```

**Step 4: 執行測試確認通過**

```bash
go test ./internal/ai/ -v
```
Expected: 全部 PASS

**Step 5: Commit**

```bash
git add internal/ai/
git commit -m "feat(ai): Claude Code 模型和工具偵測"
```

---

## 後續 Tasks（Phase 2 規劃，本次不實作）

以下 tasks 在 Phase 1 完成後再開始：

- **Task 14:** tmux popup 整合 — 偵測 `$TMUX` 環境變數，自動選擇啟動模式
- **Task 15:** Session 操作對話框 — 更名、刪除確認、新建
- **Task 16:** 群組排序操作 — Shift+J/K 移動排序
- **Task 17:** Control mode 管道 — 零子程序效能優化
- **Task 18:** AI 摘要伺服器 — `claude stream -p` + MCP server 整合
- **Task 19:** AI 建議命名
- **Task 20:** 搜尋功能

---

## 測試執行指南

```bash
# 全部測試
make test

# 單一模組
go test ./internal/config/ -v
go test ./internal/tmux/ -v
go test ./internal/store/ -v
go test ./internal/ui/ -v
go test ./internal/ai/ -v

# 帶覆蓋率
make test-cover
```

package tmux

import "strings"

// StripANSI 移除字串中的 ANSI 逸出序列，使用 O(n) 單次掃描，不使用正規表達式。
func StripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			i++
			if i < len(s) && s[i] == '[' {
				i++
				for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
					i++
				}
				if i < len(s) {
					i++
				}
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// busyIndicators 是表示 AI 正在忙碌的文字指標。
var busyIndicators = []string{"ctrl+c to interrupt", "esc to interrupt"}

// brailleChars 是 Braille 旋轉指標字元。
var brailleChars = [...]rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// waitingIndicators 是表示等待使用者輸入的文字指標。
var waitingIndicators = []string{"Yes, allow once", "No, and tell Claude", "Continue? (Y/n)", "(Y/n)"}

// DetectStatus 根據終端內容偵測 session 狀態（第三層偵測）。
func DetectStatus(content string) SessionStatus {
	if content == "" {
		return StatusIdle
	}
	clean := StripANSI(content)

	// 檢查忙碌指標
	for _, ind := range busyIndicators {
		if strings.Contains(clean, ind) {
			return StatusRunning
		}
	}

	// 檢查 Braille 旋轉指標
	for _, r := range clean {
		for _, br := range brailleChars {
			if r == br {
				return StatusRunning
			}
		}
	}

	// 檢查星號旋轉指標模式
	if containsSpinnerPattern(clean) {
		return StatusRunning
	}

	// 檢查等待指標
	for _, ind := range waitingIndicators {
		if strings.Contains(clean, ind) {
			return StatusWaiting
		}
	}

	// 檢查最後一行是否為提示字元
	if hasPromptChar(clean) {
		return StatusWaiting
	}

	return StatusIdle
}

// containsSpinnerPattern 檢查是否包含星號旋轉指標模式。
func containsSpinnerPattern(s string) bool {
	return strings.Contains(s, "* Clauding") || strings.Contains(s, "* Hullaballooing") ||
		strings.Contains(s, "* Thinking") || strings.Contains(s, "* Pondering")
}

// hasPromptChar 檢查最後一行是否為提示字元。
func hasPromptChar(s string) bool {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) == 0 {
		return false
	}
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	return lastLine == ">" || lastLine == "❯"
}

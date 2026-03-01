package ai

import (
	"regexp"
	"strings"
)

var claudeModelPattern = regexp.MustCompile(`claude-(?:sonnet|opus|haiku)-[\w.-]+`)

func DetectModel(content string) string {
	return claudeModelPattern.FindString(content)
}

func DetectTool(content string) string {
	claudeIndicators := []string{
		"ctrl+c to interrupt",
		"esc to interrupt",
	}
	for _, indicator := range claudeIndicators {
		if strings.Contains(content, indicator) {
			return "claude-code"
		}
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) > 0 {
		last := strings.TrimSpace(lines[len(lines)-1])
		if last == ">" || last == "\u2771" {
			return "claude-code"
		}
	}
	return ""
}

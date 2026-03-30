package kiro

import (
	"regexp"
	"strings"
)

var internalMetadataBlockPattern = regexp.MustCompile(`(?is)<environment_details>.*?</environment_details>`)

func sanitizePromptText(text string) string {
	return collapseBlankLines(stripInternalMetadataBlocks(text))
}

func sanitizeModelOutputText(text string) string {
	return collapseBlankLines(stripInternalMetadataBlocks(text))
}

func stripInternalMetadataBlocks(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	cleaned := internalMetadataBlockPattern.ReplaceAllString(trimmed, "")
	return strings.TrimSpace(cleaned)
}

func collapseBlankLines(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	parts := make([]string, 0, len(lines))
	blankPending := false
	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(line) == "" {
			if len(parts) > 0 {
				blankPending = true
			}
			continue
		}
		if blankPending {
			parts = append(parts, "")
			blankPending = false
		}
		parts = append(parts, line)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

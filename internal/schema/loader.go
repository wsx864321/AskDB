package schema

import (
	"fmt"
	"os"
	"strings"
)

func LoadPromptSchema(path string, maxBytes int) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read schema sql: %w", err)
	}
	if len(b) <= maxBytes {
		return string(b), nil
	}
	// Keep head + tail to preserve table definitions while staying in token budget.
	head := b[:maxBytes*3/4]
	tail := b[len(b)-maxBytes/4:]
	return strings.TrimSpace(string(head)) + "\n\n-- ... truncated ...\n\n" + strings.TrimSpace(string(tail)), nil
}

package schema

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Example struct {
	Question string `json:"question"`
	SQL      string `json:"sql"`
}

func LoadOptionalText(path string, maxBytes int) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read optional text %s: %w", path, err)
	}
	if len(b) > maxBytes {
		b = b[:maxBytes]
	}
	return strings.TrimSpace(string(b)), nil
}

func LoadFewShot(path string, maxBytes int) ([]Example, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open fewshot file %s: %w", path, err)
	}
	defer f.Close()

	res := make([]Example, 0)
	scanner := bufio.NewScanner(f)
	total := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		total += len(line)
		if total > maxBytes {
			break
		}
		var ex Example
		if err := json.Unmarshal([]byte(line), &ex); err != nil {
			return nil, fmt.Errorf("parse fewshot line: %w", err)
		}
		if ex.Question == "" || ex.SQL == "" {
			continue
		}
		res = append(res, ex)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

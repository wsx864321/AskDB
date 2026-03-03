package sqlguard

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var forbidden = regexp.MustCompile(`(?i)\b(insert|update|delete|drop|alter|truncate|create|replace|grant|revoke|call|execute|set|use|show)\b`)

func EnforceReadOnly(sql string, defaultLimit, maxLimit int) (string, error) {
	s := strings.TrimSpace(sql)
	s = strings.TrimSuffix(s, ";")
	if s == "" {
		return "", errors.New("empty sql")
	}
	if strings.Contains(s, ";") {
		return "", errors.New("multiple statements are not allowed")
	}
	lower := strings.ToLower(strings.TrimSpace(s))
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return "", errors.New("only SELECT/CTE queries are allowed")
	}
	if forbidden.MatchString(lower) {
		return "", errors.New("forbidden keyword detected")
	}

	if !strings.Contains(lower, " limit ") {
		s = fmt.Sprintf("%s LIMIT %d", s, defaultLimit)
		return s, nil
	}

	// Soft check for explicit numeric limit upper bound.
	idx := strings.LastIndex(lower, " limit ")
	if idx >= 0 {
		part := strings.TrimSpace(lower[idx+7:])
		n := 0
		_, err := fmt.Sscanf(part, "%d", &n)
		if err == nil && n > maxLimit {
			s = s[:idx] + fmt.Sprintf(" LIMIT %d", maxLimit)
		}
	}
	return s, nil
}

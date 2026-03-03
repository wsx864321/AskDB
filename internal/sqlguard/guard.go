package sqlguard

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var forbidden = regexp.MustCompile(`(?i)\b(insert|update|delete|drop|alter|truncate|create|replace|grant|revoke|call|execute|set|use|show|outfile|load_file|into)\b`)

func EnforceReadOnly(sql string, defaultLimit, maxLimit int) (string, error) {
	s := normalizeSQL(sql)
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
		s = fmt.Sprintf("%s LIMIT %d", strings.TrimSpace(s), defaultLimit)
		return s, nil
	}

	idx := strings.LastIndex(lower, " limit ")
	if idx >= 0 {
		part := strings.TrimSpace(lower[idx+7:])
		n := 0
		_, err := fmt.Sscanf(part, "%d", &n)
		if err == nil && n > maxLimit {
			s = strings.TrimSpace(s[:idx]) + fmt.Sprintf(" LIMIT %d", maxLimit)
		}
	}
	return s, nil
}

func normalizeSQL(sql string) string {
	s := strings.TrimSpace(sql)
	s = strings.TrimSuffix(s, ";")
	lineParts := strings.Split(s, "\n")
	clean := make([]string, 0, len(lineParts))
	for _, line := range lineParts {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "--") || strings.HasPrefix(l, "#") {
			continue
		}
		if i := strings.Index(l, "--"); i >= 0 {
			l = strings.TrimSpace(l[:i])
		}
		if l != "" {
			clean = append(clean, l)
		}
	}
	return strings.Join(clean, " ")
}

package schema

import (
	"regexp"
	"strings"
)

type TableBlock struct {
	Name string
	SQL  string
}

var createTableRegexp = regexp.MustCompile(`(?is)create\s+table\s+` + "`?" + `([a-zA-Z0-9_]+)` + "`?" + `\s*\((.*?)\)\s*[^;]*;`)

func ExtractTableBlocks(schemaSQL string) []TableBlock {
	matches := createTableRegexp.FindAllStringSubmatch(schemaSQL, -1)
	res := make([]TableBlock, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		name := strings.TrimSpace(m[1])
		sql := strings.TrimSpace(m[0])
		if name == "" || sql == "" {
			continue
		}
		res = append(res, TableBlock{Name: strings.ToLower(name), SQL: sql})
	}
	return res
}

package retrieval

import (
	"regexp"
	"sort"
	"strings"

	"askdb/internal/schema"
)

type Retriever struct {
	tables []schema.TableBlock
}

func New(schemaSQL string) *Retriever {
	return &Retriever{tables: schema.ExtractTableBlocks(schemaSQL)}
}

func (r *Retriever) BuildPromptSchema(question, glossary string, topK, maxBytes int) string {
	if len(r.tables) == 0 {
		return ""
	}
	if topK <= 0 || topK > len(r.tables) {
		topK = len(r.tables)
	}
	qTokens := tokenize(question + " " + glossary)

	type scored struct {
		t schema.TableBlock
		s int
	}
	scores := make([]scored, 0, len(r.tables))
	for _, tb := range r.tables {
		text := strings.ToLower(tb.Name + " " + tb.SQL)
		toks := tokenize(text)
		set := make(map[string]struct{}, len(toks))
		for _, tk := range toks {
			set[tk] = struct{}{}
		}
		s := 0
		for _, q := range qTokens {
			if _, ok := set[q]; ok {
				s += 3
			}
		}
		if strings.Contains(strings.ToLower(question), tb.Name) {
			s += 8
		}
		scores = append(scores, scored{t: tb, s: s})
	}

	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].s == scores[j].s {
			return scores[i].t.Name < scores[j].t.Name
		}
		return scores[i].s > scores[j].s
	})

	picked := make([]string, 0, topK+1)
	total := 0
	for i := 0; i < len(scores) && len(picked) < topK; i++ {
		block := scores[i].t.SQL
		if scores[i].s == 0 && len(picked) > 0 {
			break
		}
		if total+len(block)+2 > maxBytes && len(picked) > 0 {
			break
		}
		picked = append(picked, block)
		total += len(block) + 2
	}

	if len(picked) == 0 {
		// fallback: include first N tables by size constraints.
		for _, tb := range r.tables {
			if total+len(tb.SQL)+2 > maxBytes && len(picked) > 0 {
				break
			}
			picked = append(picked, tb.SQL)
			total += len(tb.SQL) + 2
			if len(picked) >= topK {
				break
			}
		}
	}
	return strings.Join(picked, "\n\n")
}

var tokenRe = regexp.MustCompile(`[a-zA-Z0-9_]{2,}`)

func tokenize(s string) []string {
	m := tokenRe.FindAllString(strings.ToLower(s), -1)
	return m
}

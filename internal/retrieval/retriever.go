package retrieval

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"askdb/internal/schema"
)

type Retriever struct {
	tables []tableDoc
	avgDL  float64
	N      float64
}

type tableDoc struct {
	block  schema.TableBlock
	tf     map[string]float64
	dl     float64
	nameLC string
}

type Options struct {
	TopK         int
	MaxBytes     int
	KeywordBoost float64
	BM25Weight   float64
	NameBoost    float64
}

func New(schemaSQL string) *Retriever {
	blocks := schema.ExtractTableBlocks(schemaSQL)
	tables := make([]tableDoc, 0, len(blocks))
	totalDL := 0.0
	for _, b := range blocks {
		toks := tokenize(strings.ToLower(b.Name + " " + b.SQL))
		tf := make(map[string]float64, len(toks))
		for _, tk := range toks {
			tf[tk] += 1
		}
		dl := float64(len(toks))
		totalDL += dl
		tables = append(tables, tableDoc{block: b, tf: tf, dl: dl, nameLC: strings.ToLower(b.Name)})
	}
	avg := 1.0
	if len(tables) > 0 {
		avg = totalDL / float64(len(tables))
	}
	return &Retriever{tables: tables, avgDL: avg, N: float64(len(tables))}
}

func (r *Retriever) BuildPromptSchema(question, glossary string, opts Options) string {
	if len(r.tables) == 0 {
		return ""
	}
	if opts.TopK <= 0 || opts.TopK > len(r.tables) {
		opts.TopK = len(r.tables)
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 60000
	}
	if opts.KeywordBoost <= 0 {
		opts.KeywordBoost = 1
	}
	if opts.BM25Weight <= 0 {
		opts.BM25Weight = 1
	}
	if opts.NameBoost <= 0 {
		opts.NameBoost = 8
	}

	qTokens := tokenize(question + " " + glossary)
	df := r.docFreq(qTokens)

	type scored struct {
		t schema.TableBlock
		s float64
	}
	scores := make([]scored, 0, len(r.tables))
	for _, tb := range r.tables {
		tokenHit := 0.0
		for _, q := range qTokens {
			if tb.tf[q] > 0 {
				tokenHit += 1
			}
		}
		bm25 := r.bm25(tb, qTokens, df)
		s := opts.KeywordBoost*tokenHit + opts.BM25Weight*bm25
		if strings.Contains(strings.ToLower(question), tb.nameLC) {
			s += opts.NameBoost
		}
		scores = append(scores, scored{t: tb.block, s: s})
	}

	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].s == scores[j].s {
			return scores[i].t.Name < scores[j].t.Name
		}
		return scores[i].s > scores[j].s
	})

	picked := make([]string, 0, opts.TopK+1)
	total := 0
	for i := 0; i < len(scores) && len(picked) < opts.TopK; i++ {
		block := scores[i].t.SQL
		if scores[i].s <= 0 && len(picked) > 0 {
			break
		}
		if total+len(block)+2 > opts.MaxBytes && len(picked) > 0 {
			break
		}
		picked = append(picked, block)
		total += len(block) + 2
	}

	if len(picked) == 0 {
		for _, tb := range r.tables {
			if total+len(tb.block.SQL)+2 > opts.MaxBytes && len(picked) > 0 {
				break
			}
			picked = append(picked, tb.block.SQL)
			total += len(tb.block.SQL) + 2
			if len(picked) >= opts.TopK {
				break
			}
		}
	}
	return strings.Join(picked, "\n\n")
}

func (r *Retriever) docFreq(tokens []string) map[string]float64 {
	uniq := map[string]struct{}{}
	for _, t := range tokens {
		uniq[t] = struct{}{}
	}
	df := make(map[string]float64, len(uniq))
	for t := range uniq {
		for _, d := range r.tables {
			if d.tf[t] > 0 {
				df[t] += 1
			}
		}
	}
	return df
}

func (r *Retriever) bm25(d tableDoc, qTokens []string, df map[string]float64) float64 {
	const k1 = 1.2
	const b = 0.75
	score := 0.0
	for _, q := range qTokens {
		tf := d.tf[q]
		if tf <= 0 {
			continue
		}
		nq := df[q]
		idf := math.Log(1 + (r.N-nq+0.5)/(nq+0.5))
		den := tf + k1*(1-b+b*(d.dl/r.avgDL))
		score += idf * ((tf * (k1 + 1)) / den)
	}
	return score
}

var tokenRe = regexp.MustCompile(`[a-zA-Z0-9_]{2,}`)

func tokenize(s string) []string {
	return tokenRe.FindAllString(strings.ToLower(s), -1)
}

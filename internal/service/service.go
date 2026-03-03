package service

import (
	"context"
	"fmt"

	"askdb/internal/llm"
	"askdb/internal/retrieval"
	"askdb/internal/schema"
	"askdb/internal/sqlguard"
)

type Service struct {
	llm              *llm.Client
	schemaText       string
	glossary         string
	examples         []schema.Example
	defaultRowLimit  int
	maxRowLimit      int
	guardRepairTries int
	recallTopK       int
	recallMaxBytes   int
	recallBM25Weight float64
	recallLexWeight  float64
	recallNameBoost  float64
	retriever        *retrieval.Retriever
}

type Request struct {
	Question string `json:"question"`
	MaxRows  int    `json:"max_rows"`
	Execute  bool   `json:"execute"`
}

type Response struct {
	SQL            string `json:"sql"`
	Reasoning      string `json:"reasoning,omitempty"`
	Attempts       int    `json:"attempts"`
	RecalledTables int    `json:"recalled_tables"`
}

func New(llmClient *llm.Client, schemaText, glossary string, examples []schema.Example, defaultLimit, maxLimit, guardRepairTries, recallTopK, recallMaxBytes int, recallBM25Weight, recallLexWeight, recallNameBoost float64) *Service {
	return &Service{
		llm:              llmClient,
		schemaText:       schemaText,
		glossary:         glossary,
		examples:         examples,
		defaultRowLimit:  defaultLimit,
		maxRowLimit:      maxLimit,
		guardRepairTries: guardRepairTries,
		recallTopK:       recallTopK,
		recallMaxBytes:   recallMaxBytes,
		recallBM25Weight: recallBM25Weight,
		recallLexWeight:  recallLexWeight,
		recallNameBoost:  recallNameBoost,
		retriever:        retrieval.New(schemaText),
	}
}

func (s *Service) Generate(ctx context.Context, req Request) (*Response, error) {
	if req.Question == "" {
		return nil, fmt.Errorf("question is required")
	}
	if req.Execute {
		return nil, fmt.Errorf("execute=true is not enabled in this build; this API only returns read-only SQL")
	}

	rowLimit := req.MaxRows
	if rowLimit <= 0 {
		rowLimit = s.defaultRowLimit
	}
	if rowLimit > s.maxRowLimit {
		rowLimit = s.maxRowLimit
	}

	recalledSchema := s.retriever.BuildPromptSchema(req.Question, s.glossary, retrieval.Options{
		TopK:         s.recallTopK,
		MaxBytes:     s.recallMaxBytes,
		KeywordBoost: s.recallLexWeight,
		BM25Weight:   s.recallBM25Weight,
		NameBoost:    s.recallNameBoost,
	})
	if recalledSchema == "" {
		recalledSchema = s.schemaText
	}
	recalledCount := len(schema.ExtractTableBlocks(recalledSchema))

	gen, err := s.llm.GenerateSQL(ctx, recalledSchema, s.glossary, req.Question, s.examples, rowLimit)
	if err != nil {
		return nil, err
	}

	attempts := 1
	sqlCandidate := gen.SQL
	reason := gen.Reasoning
	for {
		safeSQL, guardErr := sqlguard.EnforceReadOnly(sqlCandidate, rowLimit, s.maxRowLimit)
		if guardErr == nil {
			return &Response{SQL: safeSQL, Reasoning: reason, Attempts: attempts, RecalledTables: recalledCount}, nil
		}
		if attempts > s.guardRepairTries {
			return nil, fmt.Errorf("sql guard failed after %d attempts: %w", attempts, guardErr)
		}
		fix, err := s.llm.RepairSQL(ctx, recalledSchema, s.glossary, req.Question, sqlCandidate, guardErr.Error(), rowLimit)
		if err != nil {
			return nil, fmt.Errorf("sql guard failed (%v) and repair failed: %w", guardErr, err)
		}
		sqlCandidate = fix.SQL
		reason = fix.Reasoning
		attempts++
	}
}

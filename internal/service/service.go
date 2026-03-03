package service

import (
	"context"
	"fmt"

	"askdb/internal/llm"
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
}

type Request struct {
	Question string `json:"question"`
	MaxRows  int    `json:"max_rows"`
	Execute  bool   `json:"execute"`
}

type Response struct {
	SQL       string `json:"sql"`
	Reasoning string `json:"reasoning,omitempty"`
	Attempts  int    `json:"attempts"`
}

func New(llmClient *llm.Client, schemaText, glossary string, examples []schema.Example, defaultLimit, maxLimit, guardRepairTries int) *Service {
	return &Service{
		llm:              llmClient,
		schemaText:       schemaText,
		glossary:         glossary,
		examples:         examples,
		defaultRowLimit:  defaultLimit,
		maxRowLimit:      maxLimit,
		guardRepairTries: guardRepairTries,
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

	gen, err := s.llm.GenerateSQL(ctx, s.schemaText, s.glossary, req.Question, s.examples, rowLimit)
	if err != nil {
		return nil, err
	}

	attempts := 1
	sqlCandidate := gen.SQL
	reason := gen.Reasoning
	for {
		safeSQL, guardErr := sqlguard.EnforceReadOnly(sqlCandidate, rowLimit, s.maxRowLimit)
		if guardErr == nil {
			return &Response{SQL: safeSQL, Reasoning: reason, Attempts: attempts}, nil
		}
		if attempts > s.guardRepairTries {
			return nil, fmt.Errorf("sql guard failed after %d attempts: %w", attempts, guardErr)
		}
		fix, err := s.llm.RepairSQL(ctx, s.schemaText, s.glossary, req.Question, sqlCandidate, guardErr.Error(), rowLimit)
		if err != nil {
			return nil, fmt.Errorf("sql guard failed (%v) and repair failed: %w", guardErr, err)
		}
		sqlCandidate = fix.SQL
		reason = fix.Reasoning
		attempts++
	}
}

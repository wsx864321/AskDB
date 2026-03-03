package service

import (
	"context"
	"fmt"

	"askdb/internal/llm"
	"askdb/internal/sqlguard"
)

type Service struct {
	llm             *llm.Client
	schemaText      string
	defaultRowLimit int
	maxRowLimit     int
}

type Request struct {
	Question string `json:"question"`
	MaxRows  int    `json:"max_rows"`
	Execute  bool   `json:"execute"`
}

type Response struct {
	SQL       string `json:"sql"`
	Reasoning string `json:"reasoning,omitempty"`
}

func New(llmClient *llm.Client, schemaText string, defaultLimit, maxLimit int) *Service {
	return &Service{llm: llmClient, schemaText: schemaText, defaultRowLimit: defaultLimit, maxRowLimit: maxLimit}
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

	gen, err := s.llm.GenerateSQL(ctx, s.schemaText, req.Question, rowLimit)
	if err != nil {
		return nil, err
	}
	safeSQL, err := sqlguard.EnforceReadOnly(gen.SQL, rowLimit, s.maxRowLimit)
	if err != nil {
		return nil, err
	}

	return &Response{SQL: safeSQL, Reasoning: gen.Reasoning}, nil
}

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"askdb/internal/schema"
)

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type Generation struct {
	SQL       string `json:"sql"`
	Reasoning string `json:"reasoning"`
}

func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) GenerateSQL(ctx context.Context, schemaText, glossary, question string, examples []schema.Example, rowLimit int) (*Generation, error) {
	userPrompt := buildGenerationPrompt(schemaText, glossary, question, examples, rowLimit)
	return c.chatJSON(ctx, "你是一个MySQL只读SQL生成器。仅根据给定schema生成SQL。", userPrompt)
}

func (c *Client) RepairSQL(ctx context.Context, schemaText, glossary, question, previousSQL, violation string, rowLimit int) (*Generation, error) {
	userPrompt := fmt.Sprintf("请修复上一个SQL以满足只读安全限制。\n\nSchema:\n%s\n\nGlossary:\n%s\n\n问题:\n%s\n\n之前SQL:\n%s\n\n违反规则:\n%s\n\n输出严格JSON: {\"sql\":\"...\",\"reasoning\":\"...\"}。保持业务语义不变，并确保LIMIT<=%d。",
		schemaText, glossary, question, previousSQL, violation, rowLimit)
	return c.chatJSON(ctx, "你是一个MySQL只读SQL修复器。", userPrompt)
}

func (c *Client) chatJSON(ctx context.Context, systemPrompt, userPrompt string) (*Generation, error) {
	reqBody := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai error status=%d body=%s", resp.StatusCode, string(body))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("parse openai response: %w", err)
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("empty choices from openai")
	}

	content := extractJSON(out.Choices[0].Message.Content)
	var gen Generation
	if err := json.Unmarshal([]byte(content), &gen); err != nil {
		return nil, fmt.Errorf("parse generation json: %w content=%s", err, out.Choices[0].Message.Content)
	}
	return &gen, nil
}

func buildGenerationPrompt(schemaText, glossary, question string, examples []schema.Example, rowLimit int) string {
	parts := []string{fmt.Sprintf("Schema:\n%s", schemaText)}
	if strings.TrimSpace(glossary) != "" {
		parts = append(parts, fmt.Sprintf("Glossary:\n%s", glossary))
	}
	if len(examples) > 0 {
		b := strings.Builder{}
		b.WriteString("Few-shot examples:\n")
		for i, ex := range examples {
			b.WriteString(fmt.Sprintf("%d) Q: %s\nSQL: %s\n", i+1, ex.Question, ex.SQL))
		}
		parts = append(parts, b.String())
	}
	parts = append(parts, fmt.Sprintf("用户问题:\n%s", question))
	parts = append(parts, fmt.Sprintf("规则: 仅MySQL SELECT/CTE；禁止写操作；无明确要求时附加LIMIT；LIMIT不超过%d。", rowLimit))
	parts = append(parts, "输出严格JSON: {\"sql\":\"...\",\"reasoning\":\"...\"}")
	return strings.Join(parts, "\n\n")
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

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

func (c *Client) GenerateSQL(ctx context.Context, schemaText, question string, rowLimit int) (*Generation, error) {
	systemPrompt := "你是一个MySQL只读SQL生成器。仅根据给定schema生成SQL。必须遵守：1) 只能生成SELECT/CTE查询；2) 不得生成写操作；3) 使用MySQL语法；4) 没有明确要求时结果限制LIMIT。输出严格JSON: {\"sql\":\"...\",\"reasoning\":\"...\"}"
	userPrompt := fmt.Sprintf("schema:\n%s\n\n用户问题:\n%s\n\n默认LIMIT=%d，请生成最准确SQL。", schemaText, question, rowLimit)

	reqBody := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0,
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

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

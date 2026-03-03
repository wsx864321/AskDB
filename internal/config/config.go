package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             string
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	OpenAIModel      string
	SchemaSQLPath    string
	GlossaryPath     string
	FewShotPath      string
	PromptMaxBytes   int
	DefaultRowLimit  int
	MaxRowLimit      int
	GuardRepairTries int
	RecallTopK       int
	RecallMaxBytes   int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:    getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:      getEnv("OPENAI_MODEL", "gpt-4.1-mini"),
		SchemaSQLPath:    getEnv("SCHEMA_SQL_PATH", "schema.sql"),
		GlossaryPath:     getEnv("GLOSSARY_PATH", "glossary.md"),
		FewShotPath:      getEnv("FEWSHOT_PATH", "fewshot.jsonl"),
		PromptMaxBytes:   getEnvInt("PROMPT_MAX_BYTES", 120000),
		DefaultRowLimit:  getEnvInt("DEFAULT_ROW_LIMIT", 100),
		MaxRowLimit:      getEnvInt("MAX_ROW_LIMIT", 1000),
		GuardRepairTries: getEnvInt("GUARD_REPAIR_TRIES", 2),
		RecallTopK:       getEnvInt("RECALL_TOP_K", 12),
		RecallMaxBytes:   getEnvInt("RECALL_MAX_BYTES", 60000),
	}

	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}
	if cfg.DefaultRowLimit <= 0 || cfg.MaxRowLimit <= 0 || cfg.DefaultRowLimit > cfg.MaxRowLimit {
		return nil, fmt.Errorf("invalid row limit settings")
	}
	if cfg.GuardRepairTries < 0 || cfg.GuardRepairTries > 5 {
		return nil, fmt.Errorf("GUARD_REPAIR_TRIES must be in [0,5]")
	}
	if cfg.RecallTopK <= 0 {
		return nil, fmt.Errorf("RECALL_TOP_K must be > 0")
	}
	if cfg.RecallMaxBytes <= 0 {
		return nil, fmt.Errorf("RECALL_MAX_BYTES must be > 0")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

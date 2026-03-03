package main

import (
	"log"
	"net/http"
	"time"

	"askdb/internal/config"
	"askdb/internal/httpapi"
	"askdb/internal/llm"
	"askdb/internal/schema"
	"askdb/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	schemaText, err := schema.LoadPromptSchema(cfg.SchemaSQLPath, cfg.PromptMaxBytes)
	if err != nil {
		log.Fatalf("load schema: %v", err)
	}

	llmClient := llm.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)
	svc := service.New(llmClient, schemaText, cfg.DefaultRowLimit, cfg.MaxRowLimit)
	h := httpapi.NewHandler(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("nl2sql server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

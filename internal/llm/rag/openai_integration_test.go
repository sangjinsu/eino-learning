package rag

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
)

func TestOpenAIRAGIntegration(t *testing.T) {
	if !llmopenai.IntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI RAG integration test")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI RAG integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewChatModel returned error: %v", err)
	}

	service, err := NewService(ctx, NewInMemoryRetriever([]*schema.Document{
		{
			ID:      "rag",
			Content: "Chapter 09 explains retrieval augmented generation with an in-memory keyword retriever and source metadata.",
			MetaData: map[string]any{
				MetaKeyTitle:  "Chapter 09 RAG Basics",
				MetaKeySource: "integration-fixture",
			},
		},
	}), chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	result, err := service.Ask(ctx, "한 문장으로 Chapter 09 RAG는 무엇을 설명하나요?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if strings.TrimSpace(result.Answer) == "" {
		t.Fatal("answer is blank")
	}
	if len(result.Sources) != 1 {
		t.Fatalf("sources length = %d, want 1", len(result.Sources))
	}
	if result.Sources[0].ID != "rag" {
		t.Fatalf("source ID = %q, want rag", result.Sources[0].ID)
	}
	if len(result.PromptMessages) == 0 {
		t.Fatal("prompt messages are empty")
	}
}

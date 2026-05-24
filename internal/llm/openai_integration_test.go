package llm

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestOpenAIChatModelIntegration(t *testing.T) {
	if !OpenAIIntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI integration test")
	}

	cfg := LoadOpenAIConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := NewOpenAIChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOpenAIChatModel returned error: %v", err)
	}

	service := NewChatService(chatModel)
	answer, err := service.Ask(ctx, "In one short sentence, what does Eino ChatModel do?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if strings.TrimSpace(answer) == "" {
		t.Fatal("answer is blank")
	}
}

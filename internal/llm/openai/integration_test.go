package openai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sangjinsu/eino-learning/internal/llm/chat"
)

func TestOpenAIChatModelIntegration(t *testing.T) {
	if !IntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI integration test")
	}

	cfg := LoadConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := NewChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewChatModel returned error: %v", err)
	}

	service := chat.NewService(chatModel)
	answer, err := service.Ask(ctx, "In one short sentence, what does Eino ChatModel do?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if strings.TrimSpace(answer) == "" {
		t.Fatal("answer is blank")
	}
}

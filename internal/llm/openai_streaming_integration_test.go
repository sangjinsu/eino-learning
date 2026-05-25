package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

func TestOpenAIChatStreamingIntegration(t *testing.T) {
	if !OpenAIIntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI streaming integration test")
	}

	cfg := LoadOpenAIConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI streaming integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := NewOpenAIChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOpenAIChatModel returned error: %v", err)
	}

	service := NewChatService(chatModel)
	result, err := service.AskStreamingWithHistory(ctx, "In one short sentence, what does Eino streaming provide?", []*schema.Message{
		schema.UserMessage("What did Chapter 6 cover?"),
		schema.AssistantMessage("It covered Graph branching.", nil),
	})
	if err != nil {
		t.Fatalf("AskStreamingWithHistory returned error: %v", err)
	}
	if strings.TrimSpace(result.Answer) == "" {
		t.Fatal("answer is blank")
	}
	if len(result.Chunks) == 0 {
		t.Fatal("stream chunks are empty")
	}
	if len(result.PromptMessages) == 0 {
		t.Fatal("prompt messages are empty")
	}
}

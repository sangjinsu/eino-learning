package graph

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
)

func TestOpenAIAssistantGraphIntegration(t *testing.T) {
	if !llmopenai.IntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI graph integration test")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI graph integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewChatModel returned error: %v", err)
	}
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	calculation, err := service.Run(ctx, Input{Question: "calculate: 12 * (7 + 3)"})
	if err != nil {
		t.Fatalf("calculation Run returned error: %v", err)
	}
	if calculation.Route != RouteCalculator {
		t.Fatalf("calculation route = %q, want %q", calculation.Route, RouteCalculator)
	}
	if calculation.Answer != "12 * (7 + 3) = 120" {
		t.Fatalf("calculation answer = %q, want 120", calculation.Answer)
	}

	chat, err := service.Run(ctx, Input{
		Question: "In one short sentence, how is Eino Graph different from Chain?",
		History: []*schema.Message{
			schema.UserMessage("What did Chapter 5 cover?"),
			schema.AssistantMessage("It covered Chain.", nil),
		},
	})
	if err != nil {
		t.Fatalf("chat Run returned error: %v", err)
	}
	if chat.Route != RouteChat {
		t.Fatalf("chat route = %q, want %q", chat.Route, RouteChat)
	}
	if strings.TrimSpace(chat.Answer) == "" {
		t.Fatal("chat answer is blank")
	}
	if len(chat.PromptMessages) == 0 {
		t.Fatal("chat prompt messages are empty")
	}
}

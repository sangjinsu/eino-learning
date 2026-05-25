package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

func TestOpenAIAssistantGraphIntegration(t *testing.T) {
	if !OpenAIIntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI graph integration test")
	}

	cfg := LoadOpenAIConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI graph integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chatModel, err := NewOpenAIChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOpenAIChatModel returned error: %v", err)
	}
	service, err := NewAssistantGraphService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewAssistantGraphService returned error: %v", err)
	}

	calculation, err := service.Run(ctx, AssistantGraphInput{Question: "calculate: 12 * (7 + 3)"})
	if err != nil {
		t.Fatalf("calculation Run returned error: %v", err)
	}
	if calculation.Route != GraphRouteCalculator {
		t.Fatalf("calculation route = %q, want %q", calculation.Route, GraphRouteCalculator)
	}
	if calculation.Answer != "12 * (7 + 3) = 120" {
		t.Fatalf("calculation answer = %q, want 120", calculation.Answer)
	}

	chat, err := service.Run(ctx, AssistantGraphInput{
		Question: "In one short sentence, how is Eino Graph different from Chain?",
		History: []*schema.Message{
			schema.UserMessage("What did Chapter 5 cover?"),
			schema.AssistantMessage("It covered Chain.", nil),
		},
	})
	if err != nil {
		t.Fatalf("chat Run returned error: %v", err)
	}
	if chat.Route != GraphRouteChat {
		t.Fatalf("chat route = %q, want %q", chat.Route, GraphRouteChat)
	}
	if strings.TrimSpace(chat.Answer) == "" {
		t.Fatal("chat answer is blank")
	}
	if len(chat.PromptMessages) == 0 {
		t.Fatal("chat prompt messages are empty")
	}
}

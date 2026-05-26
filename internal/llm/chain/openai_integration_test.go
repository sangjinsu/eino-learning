package chain

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
)

func TestOpenAIChatChainIntegration(t *testing.T) {
	if !llmopenai.IntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI chain integration test")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI chain integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewChatModel returned error: %v", err)
	}

	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	answer, err := service.AskWithHistory(ctx, "In one short sentence, what does Eino Chain do?", []*schema.Message{
		schema.UserMessage("What did the previous chapter cover?"),
		schema.AssistantMessage("It covered model-backed tool calling.", nil),
	})
	if err != nil {
		t.Fatalf("AskWithHistory returned error: %v", err)
	}
	if strings.TrimSpace(answer) == "" {
		t.Fatal("answer is blank")
	}
}

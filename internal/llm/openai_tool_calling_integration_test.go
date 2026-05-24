package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestOpenAIToolCallingIntegration(t *testing.T) {
	if !OpenAIIntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI tool calling integration test")
	}

	cfg := LoadOpenAIConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI tool calling integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chatModel, err := NewOpenAIChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOpenAIChatModel returned error: %v", err)
	}
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}

	service := NewChatService(chatModel)
	result, err := service.AskWithTools(
		ctx,
		`Use the calculator tool to calculate "12 * (7 + 3)", then answer in one short sentence.`,
		[]einotool.BaseTool{calculatorTool},
	)
	if err != nil {
		t.Fatalf("AskWithTools returned error: %v", err)
	}
	if len(result.FirstResponse.ToolCalls) == 0 {
		t.Fatalf("model did not request a tool call; final answer = %q", result.Answer())
	}
	if len(result.ToolMessages) == 0 {
		t.Fatal("tool messages are empty")
	}
	if strings.TrimSpace(result.Answer()) == "" {
		t.Fatal("final answer is blank")
	}
}

package toolcalling

import (
	"context"
	"strings"
	"testing"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestOpenAIToolCallingIntegration(t *testing.T) {
	if !llmopenai.IntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI tool calling integration test")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI tool calling integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewChatModel returned error: %v", err)
	}
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}

	service := NewService(chatModel)
	result, err := service.Ask(
		ctx,
		`Use the calculator tool to calculate "12 * (7 + 3)", then answer in one short sentence.`,
		[]einotool.BaseTool{calculatorTool},
	)
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
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

package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestOpenAIReActAgentIntegration(t *testing.T) {
	if !llmopenai.IntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI ReAct agent integration test")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI ReAct agent integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewChatModel returned error: %v", err)
	}
	toolCallingModel, ok := chatModel.(model.ToolCallingChatModel)
	if !ok {
		t.Fatal("OpenAI chat model does not support tool calling")
	}
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}

	service := NewService(toolCallingModel, []einotool.BaseTool{calculatorTool})
	result, err := service.Ask(ctx, `Use the calculator tool to calculate "12 * (3 + 4)", then answer in one short sentence.`)
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if strings.TrimSpace(result.Answer()) == "" {
		t.Fatal("final answer is blank")
	}
	if len(result.AvailableTools) != 1 || result.AvailableTools[0] != tools.CalculatorToolName {
		t.Fatalf("available tools = %#v, want calculator", result.AvailableTools)
	}
}

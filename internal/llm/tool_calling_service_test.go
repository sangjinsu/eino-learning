package llm

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestAskWithToolsLetsModelRequestToolAndReturnsFinalAnswer(t *testing.T) {
	ctx := context.Background()
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	state := &scriptedToolCallingState{requestTool: true}
	service := NewChatService(&scriptedToolCallingModel{state: state})

	result, err := service.AskWithTools(ctx, "Calculate 2 + 3 * 4.", []einotool.BaseTool{calculatorTool})
	if err != nil {
		t.Fatalf("AskWithTools returned error: %v", err)
	}

	if result.Answer() != "2 + 3 * 4 = 14." {
		t.Fatalf("answer = %q, want final model answer", result.Answer())
	}
	if state.withToolsCalls != 1 {
		t.Fatalf("WithTools call count = %d, want 1", state.withToolsCalls)
	}
	if len(state.boundTools) != 1 {
		t.Fatalf("bound tool count = %d, want 1", len(state.boundTools))
	}
	if state.boundTools[0].Name != tools.CalculatorToolName {
		t.Fatalf("bound tool name = %q, want %q", state.boundTools[0].Name, tools.CalculatorToolName)
	}
	if len(state.generateInputs) != 2 {
		t.Fatalf("Generate call count = %d, want 2", len(state.generateInputs))
	}
	if len(result.ToolMessages) != 1 {
		t.Fatalf("tool message count = %d, want 1", len(result.ToolMessages))
	}

	secondInput := state.generateInputs[1]
	if len(secondInput) < 2 {
		t.Fatalf("second Generate input has %d messages, want tool call and tool result", len(secondInput))
	}
	assistantMessage := secondInput[len(secondInput)-2]
	if len(assistantMessage.ToolCalls) != 1 {
		t.Fatalf("assistant tool call count = %d, want 1", len(assistantMessage.ToolCalls))
	}
	toolMessage := secondInput[len(secondInput)-1]
	if toolMessage.Role != schema.Tool {
		t.Fatalf("last message role = %q, want %q", toolMessage.Role, schema.Tool)
	}
	if toolMessage.ToolCallID != "call_calculator" {
		t.Fatalf("tool call id = %q, want call_calculator", toolMessage.ToolCallID)
	}
	if !strings.Contains(toolMessage.Content, `"result":14`) {
		t.Fatalf("tool message content = %q, want calculator result JSON", toolMessage.Content)
	}
}

func TestAskWithToolsReturnsDirectAnswerWhenModelDoesNotRequestTool(t *testing.T) {
	ctx := context.Background()
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	state := &scriptedToolCallingState{requestTool: false}
	service := NewChatService(&scriptedToolCallingModel{state: state})

	result, err := service.AskWithTools(ctx, "Say hello without tools.", []einotool.BaseTool{calculatorTool})
	if err != nil {
		t.Fatalf("AskWithTools returned error: %v", err)
	}

	if result.Answer() != "No tool is needed." {
		t.Fatalf("answer = %q, want direct answer", result.Answer())
	}
	if len(result.ToolMessages) != 0 {
		t.Fatalf("tool message count = %d, want 0", len(result.ToolMessages))
	}
	if len(state.generateInputs) != 1 {
		t.Fatalf("Generate call count = %d, want 1", len(state.generateInputs))
	}
}

func TestAskWithToolsRejectsBlankQuestionBeforeBindingTools(t *testing.T) {
	ctx := context.Background()
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	state := &scriptedToolCallingState{requestTool: true}
	service := NewChatService(&scriptedToolCallingModel{state: state})

	_, err = service.AskWithTools(ctx, " ", []einotool.BaseTool{calculatorTool})
	if !errors.Is(err, ErrBlankQuestion) {
		t.Fatalf("AskWithTools error = %v, want %v", err, ErrBlankQuestion)
	}
	if state.withToolsCalls != 0 {
		t.Fatalf("WithTools call count = %d, want 0", state.withToolsCalls)
	}
	if len(state.generateInputs) != 0 {
		t.Fatalf("Generate call count = %d, want 0", len(state.generateInputs))
	}
}

type scriptedToolCallingState struct {
	requestTool    bool
	withToolsCalls int
	boundTools     []*schema.ToolInfo
	generateInputs [][]*schema.Message
}

type scriptedToolCallingModel struct {
	state *scriptedToolCallingState
}

var _ model.ToolCallingChatModel = (*scriptedToolCallingModel)(nil)

func (m *scriptedToolCallingModel) WithTools(toolInfos []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	m.state.withToolsCalls++
	m.state.boundTools = append([]*schema.ToolInfo(nil), toolInfos...)
	return &scriptedToolCallingModel{state: m.state}, nil
}

func (m *scriptedToolCallingModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	_ = ctx
	_ = opts

	m.state.generateInputs = append(m.state.generateInputs, append([]*schema.Message(nil), input...))
	if len(m.state.generateInputs) == 1 {
		if !m.state.requestTool {
			return schema.AssistantMessage("No tool is needed.", nil), nil
		}

		return schema.AssistantMessage("", []schema.ToolCall{
			{
				ID:   "call_calculator",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      tools.CalculatorToolName,
					Arguments: `{"expression":"2 + 3 * 4"}`,
				},
			},
		}), nil
	}

	return schema.AssistantMessage("2 + 3 * 4 = 14.", nil), nil
}

func (m *scriptedToolCallingModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	_ = ctx
	_ = input
	_ = opts

	return nil, errors.New("not supported in scripted test model")
}

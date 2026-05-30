package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestAskRunsReActLoopWithCalculatorTool(t *testing.T) {
	ctx := context.Background()
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	state := &scriptedReActModelState{}
	service := NewService(&scriptedReActModel{state: state}, []einotool.BaseTool{calculatorTool})

	result, err := service.Ask(ctx, "Use calculator to solve 2 + 3 * 4.")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if result.Answer() != "2 + 3 * 4 = 14." {
		t.Fatalf("answer = %q, want final answer", result.Answer())
	}
	if result.MaxStep != DefaultMaxStep {
		t.Fatalf("MaxStep = %d, want %d", result.MaxStep, DefaultMaxStep)
	}
	if len(result.AvailableTools) != 1 || result.AvailableTools[0] != tools.CalculatorToolName {
		t.Fatalf("AvailableTools = %#v, want calculator", result.AvailableTools)
	}
	if state.withToolsCalls != 1 {
		t.Fatalf("WithTools call count = %d, want 1", state.withToolsCalls)
	}
	if len(state.boundTools) != 1 || state.boundTools[0].Name != tools.CalculatorToolName {
		t.Fatalf("bound tools = %#v, want calculator tool info", state.boundTools)
	}
	if len(state.generateInputs) != 2 {
		t.Fatalf("Generate call count = %d, want 2", len(state.generateInputs))
	}

	secondInput := state.generateInputs[1]
	if len(secondInput) < 2 {
		t.Fatalf("second Generate input has %d messages, want assistant tool call and tool result", len(secondInput))
	}
	toolMessage := secondInput[len(secondInput)-1]
	if toolMessage.Role != schema.Tool {
		t.Fatalf("last message role = %q, want %q", toolMessage.Role, schema.Tool)
	}
	if toolMessage.ToolCallID != "call_calculator" {
		t.Fatalf("tool call id = %q, want call_calculator", toolMessage.ToolCallID)
	}
	if !strings.Contains(toolMessage.Content, `"result":14`) {
		t.Fatalf("tool message content = %q, want calculator JSON result", toolMessage.Content)
	}
}

func TestAskRejectsBlankQuestionBeforeBindingTools(t *testing.T) {
	ctx := context.Background()
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	state := &scriptedReActModelState{}
	service := NewService(&scriptedReActModel{state: state}, []einotool.BaseTool{calculatorTool})

	_, err = service.Ask(ctx, " \t\n ")
	if !errors.Is(err, prompting.ErrBlankQuestion) {
		t.Fatalf("Ask error = %v, want %v", err, prompting.ErrBlankQuestion)
	}
	if state.withToolsCalls != 0 {
		t.Fatalf("WithTools call count = %d, want 0", state.withToolsCalls)
	}
	if len(state.generateInputs) != 0 {
		t.Fatalf("Generate call count = %d, want 0", len(state.generateInputs))
	}
}

func TestNewServiceRejectsNilTool(t *testing.T) {
	state := &scriptedReActModelState{}
	service := NewService(&scriptedReActModel{state: state}, []einotool.BaseTool{nil})

	_, err := service.Ask(context.Background(), "Use a tool.")
	if err == nil {
		t.Fatal("Ask returned nil error for nil tool")
	}
}

func TestNewServiceWithOptionsUsesCustomMaxStep(t *testing.T) {
	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	state := &scriptedReActModelState{}
	service := NewServiceWithOptions(&scriptedReActModel{state: state}, []einotool.BaseTool{calculatorTool}, Options{
		MaxStep: 4,
	})

	result, err := service.Ask(context.Background(), "Use calculator to solve 2 + 3 * 4.")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if result.MaxStep != 4 {
		t.Fatalf("MaxStep = %d, want 4", result.MaxStep)
	}
}

type scriptedReActModelState struct {
	withToolsCalls int
	boundTools     []*schema.ToolInfo
	generateInputs [][]*schema.Message
}

type scriptedReActModel struct {
	state *scriptedReActModelState
}

var _ model.ToolCallingChatModel = (*scriptedReActModel)(nil)

func (m *scriptedReActModel) WithTools(toolInfos []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	m.state.withToolsCalls++
	m.state.boundTools = append([]*schema.ToolInfo(nil), toolInfos...)
	return &scriptedReActModel{state: m.state}, nil
}

func (m *scriptedReActModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	_ = ctx
	_ = opts

	m.state.generateInputs = append(m.state.generateInputs, append([]*schema.Message(nil), input...))
	if len(m.state.generateInputs) == 1 {
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

func (m *scriptedReActModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}

	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

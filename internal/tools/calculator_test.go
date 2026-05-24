package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func TestCalculatorToolInfo(t *testing.T) {
	calculatorTool, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}

	info, err := calculatorTool.Info(context.Background())
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	if info.Name != CalculatorToolName {
		t.Fatalf("tool name = %q, want %q", info.Name, CalculatorToolName)
	}
	if !strings.Contains(info.Desc, "arithmetic") {
		t.Fatalf("tool description = %q, want arithmetic-oriented description", info.Desc)
	}
	if info.ParamsOneOf == nil {
		t.Fatal("tool params schema is nil")
	}
	if _, err := info.ParamsOneOf.ToJSONSchema(); err != nil {
		t.Fatalf("ToJSONSchema returned error: %v", err)
	}
}

func TestCalculatorToolRunsWithJSONArguments(t *testing.T) {
	calculatorTool, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}

	got, err := calculatorTool.InvokableRun(context.Background(), `{"expression":"2 + 3 * 4"}`)
	if err != nil {
		t.Fatalf("InvokableRun returned error: %v", err)
	}

	var out CalculatorOutput
	if err := json.Unmarshal([]byte(got), &out); err != nil {
		t.Fatalf("tool output is not JSON: %v", err)
	}
	if out.Expression != "2 + 3 * 4" {
		t.Fatalf("expression = %q, want original expression", out.Expression)
	}
	if out.Result != 14 {
		t.Fatalf("result = %v, want 14", out.Result)
	}
}

func TestCalculateSupportsParenthesesAndUnaryOperators(t *testing.T) {
	got, err := Calculate(context.Background(), CalculatorInput{Expression: "-(6 - 10) / 2"})
	if err != nil {
		t.Fatalf("Calculate returned error: %v", err)
	}

	if got.Result != 2 {
		t.Fatalf("result = %v, want 2", got.Result)
	}
}

func TestCalculatorToolRejectsBlankExpression(t *testing.T) {
	calculatorTool, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}

	_, err = calculatorTool.InvokableRun(context.Background(), `{"expression":" "}`)
	if !errors.Is(err, ErrBlankExpression) {
		t.Fatalf("InvokableRun error = %v, want %v", err, ErrBlankExpression)
	}
}

func TestCalculateRejectsUnsupportedExpression(t *testing.T) {
	_, err := Calculate(context.Background(), CalculatorInput{Expression: "sqrt(4)"})
	if !errors.Is(err, ErrUnsupportedExpression) {
		t.Fatalf("Calculate error = %v, want %v", err, ErrUnsupportedExpression)
	}
}

func TestCalculateRejectsDivisionByZero(t *testing.T) {
	_, err := Calculate(context.Background(), CalculatorInput{Expression: "10 / (5 - 5)"})
	if !errors.Is(err, ErrDivisionByZero) {
		t.Fatalf("Calculate error = %v, want %v", err, ErrDivisionByZero)
	}
}

func TestExecuteToolCallsReturnsCalculatorToolMessages(t *testing.T) {
	ctx := context.Background()
	calculatorTool, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("NewCalculatorTool returned error: %v", err)
	}
	assistantMessage := schema.AssistantMessage("", []schema.ToolCall{
		{
			ID:   "call_calculator",
			Type: "function",
			Function: schema.FunctionCall{
				Name:      CalculatorToolName,
				Arguments: `{"expression":"7 * (8 + 2)"}`,
			},
		},
	})

	got, err := ExecuteToolCalls(ctx, []einotool.BaseTool{calculatorTool}, assistantMessage)
	if err != nil {
		t.Fatalf("ExecuteToolCalls returned error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("tool message count = %d, want 1", len(got))
	}
	if got[0].Role != schema.Tool {
		t.Fatalf("role = %q, want %q", got[0].Role, schema.Tool)
	}
	if got[0].ToolCallID != "call_calculator" {
		t.Fatalf("tool call id = %q, want call_calculator", got[0].ToolCallID)
	}
	if got[0].ToolName != CalculatorToolName {
		t.Fatalf("tool name = %q, want %q", got[0].ToolName, CalculatorToolName)
	}

	var out CalculatorOutput
	if err := json.Unmarshal([]byte(got[0].Content), &out); err != nil {
		t.Fatalf("tool message content is not JSON: %v", err)
	}
	if out.Expression != "7 * (8 + 2)" {
		t.Fatalf("expression = %q, want original expression", out.Expression)
	}
	if out.Result != 70 {
		t.Fatalf("result = %v, want 70", out.Result)
	}
}

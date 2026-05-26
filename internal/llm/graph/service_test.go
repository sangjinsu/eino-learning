package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

func TestAssistantGraphRoutesChatQuestionThroughModel(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("Graph uses nodes and edges.")
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	history := []*schema.Message{
		schema.UserMessage("What did Chapter 5 add?"),
		schema.AssistantMessage("It added Chain.", nil),
	}

	got, err := service.Run(ctx, Input{
		Question: "How does Graph differ from Chain?",
		History:  history,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got.Route != RouteChat {
		t.Fatalf("route = %q, want %q", got.Route, RouteChat)
	}
	if got.Answer != "Graph uses nodes and edges." {
		t.Fatalf("answer = %q, want fake model answer", got.Answer)
	}
	if got.ModelResponse == nil || got.ModelResponse.Content != "Graph uses nodes and edges." {
		t.Fatalf("model response = %#v, want fake model response", got.ModelResponse)
	}
	assertMessages(t, got.PromptMessages, []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What did Chapter 5 add?"},
		{role: schema.Assistant, content: "It added Chain."},
		{role: schema.User, content: "How does Graph differ from Chain?"},
	})
	assertMessages(t, chatModel.LastInput(), []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What did Chapter 5 add?"},
		{role: schema.Assistant, content: "It added Chain."},
		{role: schema.User, content: "How does Graph differ from Chain?"},
	})
}

func TestAssistantGraphRoutesCalculationWithoutCallingModel(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("unused")
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	got, err := service.Run(ctx, Input{
		Question: "calculate: 7 * (8 + 2)",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got.Route != RouteCalculator {
		t.Fatalf("route = %q, want %q", got.Route, RouteCalculator)
	}
	if got.Answer != "7 * (8 + 2) = 70" {
		t.Fatalf("answer = %q, want calculation answer", got.Answer)
	}
	if got.Calculation == nil {
		t.Fatal("calculation output is nil")
	}
	if got.Calculation.Expression != "7 * (8 + 2)" {
		t.Fatalf("calculation expression = %q, want original expression", got.Calculation.Expression)
	}
	if got.Calculation.Result != 70 {
		t.Fatalf("calculation result = %v, want 70", got.Calculation.Result)
	}
	if len(chatModel.LastInput()) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(chatModel.LastInput()))
	}
}

func TestAssistantGraphRejectsBlankQuestionBeforeCallingModel(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("unused")
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	_, err = service.Run(ctx, Input{Question: " \t\n "})
	if !errors.Is(err, prompting.ErrBlankQuestion) {
		t.Fatalf("Run error = %v, want %v", err, prompting.ErrBlankQuestion)
	}
	if len(chatModel.LastInput()) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(chatModel.LastInput()))
	}
}

func TestNewServiceRequiresModel(t *testing.T) {
	_, err := NewService(context.Background(), nil)
	if !errors.Is(err, ErrModelRequired) {
		t.Fatalf("NewService error = %v, want %v", err, ErrModelRequired)
	}
}

package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

func TestServiceAskUsesChainAndReturnsModelContent(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("Chain connects template output to the model.")
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	got, err := service.Ask(ctx, "What does Chapter 5 add?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if got != "Chain connects template output to the model." {
		t.Fatalf("answer = %q, want chain model response", got)
	}
	assertMessages(t, chatModel.LastInput(), []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What does Chapter 5 add?"},
	})
}

func TestServiceAskWithHistoryPreservesMessageOrder(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("History is included before the final question.")
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	history := []*schema.Message{
		schema.UserMessage("What did Chapter 4 cover?"),
		schema.AssistantMessage("It covered tool calling.", nil),
	}

	got, err := service.AskWithHistory(ctx, "What does Chain compose?", history)
	if err != nil {
		t.Fatalf("AskWithHistory returned error: %v", err)
	}

	if got != "History is included before the final question." {
		t.Fatalf("answer = %q, want chain model response", got)
	}
	assertMessages(t, chatModel.LastInput(), []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What did Chapter 4 cover?"},
		{role: schema.Assistant, content: "It covered tool calling."},
		{role: schema.User, content: "What does Chain compose?"},
	})
}

func TestServiceRejectsBlankQuestionBeforeCallingModel(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("unused")
	service, err := NewService(ctx, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	_, err = service.Ask(ctx, " \t\n ")
	if !errors.Is(err, prompting.ErrBlankQuestion) {
		t.Fatalf("Ask error = %v, want %v", err, prompting.ErrBlankQuestion)
	}
	if got := chatModel.LastInput(); len(got) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(got))
	}
}

func TestNewServiceRequiresModel(t *testing.T) {
	_, err := NewService(context.Background(), nil)
	if !errors.Is(err, ErrChainModelRequired) {
		t.Fatalf("NewService error = %v, want %v", err, ErrChainModelRequired)
	}
}

func TestRunWithTraceCapturesEachChainStage(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("Trace shows the chain stages.")
	history := []*schema.Message{
		schema.UserMessage("What did Chapter 4 cover?"),
		schema.AssistantMessage("It covered tool calling.", nil),
	}

	trace, err := RunWithTrace(ctx, chatModel, prompting.DefaultChatTemplate(), "What does Chapter 5 add?", history)
	if err != nil {
		t.Fatalf("RunWithTrace returned error: %v", err)
	}

	if trace.Answer() != "Trace shows the chain stages." {
		t.Fatalf("answer = %q, want traced model response", trace.Answer())
	}
	if trace.InputVariables["question"] != "What does Chapter 5 add?" {
		t.Fatalf("trace question = %v, want original question", trace.InputVariables["question"])
	}
	if gotHistory, ok := trace.InputVariables["history"].([]*schema.Message); !ok || len(gotHistory) != 2 {
		t.Fatalf("trace history = %#v, want two history messages", trace.InputVariables["history"])
	}
	assertMessages(t, trace.PromptMessages, []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What did Chapter 4 cover?"},
		{role: schema.Assistant, content: "It covered tool calling."},
		{role: schema.User, content: "What does Chapter 5 add?"},
	})
	if trace.ModelResponse == nil {
		t.Fatal("trace model response is nil")
	}
	if trace.ModelResponse.Content != "Trace shows the chain stages." {
		t.Fatalf("trace model response = %q, want fake response", trace.ModelResponse.Content)
	}
}

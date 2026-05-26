package observability

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

func TestRunObservableChatChainCapturesPromptAndModelEvents(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("Callback은 component 실행 과정을 관찰하게 해줍니다.")
	history := []*schema.Message{
		schema.UserMessage("Chapter 7에서는 무엇을 다뤘나요?"),
		schema.AssistantMessage("Streaming을 다뤘습니다.", nil),
	}

	got, err := RunObservableChatChain(ctx, chatModel, prompting.DefaultChatTemplate(), "Chapter 8은 무엇을 추가하나요?", history)
	if err != nil {
		t.Fatalf("RunObservableChatChain returned error: %v", err)
	}

	if got.Answer != "Callback은 component 실행 과정을 관찰하게 해줍니다." {
		t.Fatalf("answer = %q, want fake model response", got.Answer)
	}
	// Callback은 답변을 만드는 흐름을 바꾸지 않고, 옆에서 lifecycle event만 관찰합니다.
	assertCallbackTimeline(t, got.Events, []callbackEventWant{
		{timing: CallbackTimingStart, name: "", component: "Chain"},
		{timing: CallbackTimingStart, name: "ChatTemplate", component: "ChatTemplate"},
		{timing: CallbackTimingEnd, name: "ChatTemplate", component: "ChatTemplate"},
		{timing: CallbackTimingStart, name: "ChatModel", component: "ChatModel"},
		{timing: CallbackTimingEnd, name: "ChatModel", component: "ChatModel"},
		{timing: CallbackTimingEnd, name: "", component: "Chain"},
	})

	promptStart := findCallbackEvent(t, got.Events, CallbackTimingStart, "ChatTemplate")
	if !strings.Contains(promptStart.Summary, "question") {
		t.Fatalf("prompt start summary = %q, want question variable", promptStart.Summary)
	}

	promptEnd := findCallbackEvent(t, got.Events, CallbackTimingEnd, "ChatTemplate")
	if !strings.Contains(promptEnd.Summary, "messages=4") {
		t.Fatalf("prompt end summary = %q, want messages=4", promptEnd.Summary)
	}

	modelStart := findCallbackEvent(t, got.Events, CallbackTimingStart, "ChatModel")
	if !strings.Contains(modelStart.Summary, "messages=4") {
		t.Fatalf("model start summary = %q, want messages=4", modelStart.Summary)
	}

	modelEnd := findCallbackEvent(t, got.Events, CallbackTimingEnd, "ChatModel")
	if !strings.Contains(modelEnd.Summary, "Callback은 component 실행 과정을 관찰하게 해줍니다.") {
		t.Fatalf("model end summary = %q, want model response content", modelEnd.Summary)
	}

	assertMessages(t, chatModel.LastInput(), []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "Chapter 7에서는 무엇을 다뤘나요?"},
		{role: schema.Assistant, content: "Streaming을 다뤘습니다."},
		{role: schema.User, content: "Chapter 8은 무엇을 추가하나요?"},
	})
}

func TestRunObservableChatChainRejectsBlankQuestionBeforeCallingModel(t *testing.T) {
	chatModel := fake.NewChatModel("unused")

	_, err := RunObservableChatChain(context.Background(), chatModel, prompting.DefaultChatTemplate(), " \t\n ", nil)
	if !errors.Is(err, prompting.ErrBlankQuestion) {
		t.Fatalf("RunObservableChatChain error = %v, want %v", err, prompting.ErrBlankQuestion)
	}

	if got := chatModel.LastInput(); len(got) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(got))
	}
}

func TestRunObservableChatChainReturnsEventsWhenModelFails(t *testing.T) {
	ctx := context.Background()
	modelErr := errors.New("model failed")
	chatModel := &failingObservableModel{err: modelErr}

	got, err := RunObservableChatChain(ctx, chatModel, prompting.DefaultChatTemplate(), "에러가 나면 callback은 무엇을 기록하나요?", nil)
	if !errors.Is(err, modelErr) {
		t.Fatalf("RunObservableChatChain error = %v, want %v", err, modelErr)
	}
	if got != nil {
		errorEvent := findCallbackEvent(t, got.Events, CallbackTimingError, "ChatModel")
		if !strings.Contains(errorEvent.Summary, "model failed") {
			t.Fatalf("error summary = %q, want model failure", errorEvent.Summary)
		}
		return
	}

	t.Fatal("result is nil, want callback events even when model fails")
}

func findCallbackEvent(t *testing.T, events []CallbackEvent, timing CallbackTiming, name string) CallbackEvent {
	t.Helper()

	for _, event := range events {
		if event.Timing == timing && event.Name == name {
			return event
		}
	}

	t.Fatalf("event timing=%q name=%q not found in %#v", timing, name, events)
	return CallbackEvent{}
}

type callbackEventWant struct {
	timing    CallbackTiming
	name      string
	component string
}

func assertCallbackTimeline(t *testing.T, got []CallbackEvent, want []callbackEventWant) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("event length = %d, want %d: %#v", len(got), len(want), got)
	}

	for i := range want {
		if got[i].Timing != want[i].timing {
			t.Fatalf("event[%d].Timing = %q, want %q", i, got[i].Timing, want[i].timing)
		}
		if got[i].Name != want[i].name {
			t.Fatalf("event[%d].Name = %q, want %q", i, got[i].Name, want[i].name)
		}
		if got[i].Component != want[i].component {
			t.Fatalf("event[%d].Component = %q, want %q", i, got[i].Component, want[i].component)
		}
	}
}

type failingObservableModel struct {
	err error
}

var _ model.BaseChatModel = (*failingObservableModel)(nil)

func (m *failingObservableModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	_ = ctx
	_ = input
	_ = opts

	return nil, m.err
}

func (m *failingObservableModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	_ = ctx
	_ = input
	_ = opts

	return nil, m.err
}

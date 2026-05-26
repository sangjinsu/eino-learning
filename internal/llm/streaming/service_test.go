package streaming

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

func TestChatServiceAskWithHistoryCollectsChunksAndReturnsAnswer(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewStreamingChatModel("Streaming ", "returns ", "chunks.")
	service := NewService(chatModel)
	history := []*schema.Message{
		schema.UserMessage("What did Chapter 6 cover?"),
		schema.AssistantMessage("It covered Graph branching.", nil),
	}

	got, err := service.AskWithHistory(ctx, "What does Chapter 7 add?", history)
	if err != nil {
		t.Fatalf("AskWithHistory returned error: %v", err)
	}

	if got.Answer != "Streaming returns chunks." {
		t.Fatalf("answer = %q, want concatenated stream chunks", got.Answer)
	}
	assertMessages(t, got.Chunks, []messageWant{
		{role: schema.Assistant, content: "Streaming "},
		{role: schema.Assistant, content: "returns "},
		{role: schema.Assistant, content: "chunks."},
	})
	assertMessages(t, got.PromptMessages, []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What did Chapter 6 cover?"},
		{role: schema.Assistant, content: "It covered Graph branching."},
		{role: schema.User, content: "What does Chapter 7 add?"},
	})
	assertMessages(t, chatModel.LastInput(), []messageWant{
		{role: schema.System, content: prompting.DefaultSystemPrompt},
		{role: schema.User, content: "What did Chapter 6 cover?"},
		{role: schema.Assistant, content: "It covered Graph branching."},
		{role: schema.User, content: "What does Chapter 7 add?"},
	})
	if chatModel.StreamCalls() != 1 {
		t.Fatalf("StreamCalls = %d, want 1", chatModel.StreamCalls())
	}
}

func TestChatServiceStreamWithHistoryReturnsReaderForManualConsumption(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewStreamingChatModel("one", " two")
	service := NewService(chatModel)

	reader, err := service.StreamWithHistory(ctx, "How does streaming work?", nil)
	if err != nil {
		t.Fatalf("StreamWithHistory returned error: %v", err)
	}

	got, err := CollectMessageStream(reader)
	if err != nil {
		t.Fatalf("CollectMessageStream returned error: %v", err)
	}

	if got.Answer != "one two" {
		t.Fatalf("answer = %q, want collected stream content", got.Answer)
	}
	assertMessages(t, got.Chunks, []messageWant{
		{role: schema.Assistant, content: "one"},
		{role: schema.Assistant, content: " two"},
	})
}

func TestChatServiceAskRejectsBlankQuestionBeforeCallingModel(t *testing.T) {
	chatModel := fake.NewStreamingChatModel("unused")
	service := NewService(chatModel)

	_, err := service.Ask(context.Background(), " \t\n ")
	if !errors.Is(err, prompting.ErrBlankQuestion) {
		t.Fatalf("Ask error = %v, want %v", err, prompting.ErrBlankQuestion)
	}

	if got := chatModel.StreamCalls(); got != 0 {
		t.Fatalf("StreamCalls = %d, want 0", got)
	}
	if got := chatModel.LastInput(); len(got) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(got))
	}
}

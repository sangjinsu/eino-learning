package fake

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestChatModelGenerateRecordsInputAndReturnsAssistantMessage(t *testing.T) {
	ctx := context.Background()
	chatModel := NewChatModel("fake answer")
	input := []*schema.Message{
		schema.SystemMessage("You are a helpful tutor."),
		schema.UserMessage("What is Eino?"),
	}

	got, err := chatModel.Generate(ctx, input)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if got.Role != schema.Assistant {
		t.Fatalf("role = %q, want %q", got.Role, schema.Assistant)
	}
	if got.Content != "fake answer" {
		t.Fatalf("content = %q, want %q", got.Content, "fake answer")
	}

	lastInput := chatModel.LastInput()
	if len(lastInput) != len(input) {
		t.Fatalf("LastInput length = %d, want %d", len(lastInput), len(input))
	}
	if lastInput[1].Content != "What is Eino?" {
		t.Fatalf("last user content = %q, want %q", lastInput[1].Content, "What is Eino?")
	}
}

func TestChatModelGenerateRejectsEmptyInput(t *testing.T) {
	chatModel := NewChatModel("unused")

	_, err := chatModel.Generate(context.Background(), nil)
	if err == nil {
		t.Fatal("Generate returned nil error for empty input")
	}
}

func TestChatModelGenerateRejectsBlankLastUserMessage(t *testing.T) {
	chatModel := NewChatModel("unused")
	input := []*schema.Message{
		schema.UserMessage(" \t\n "),
	}

	_, err := chatModel.Generate(context.Background(), input)
	if err == nil {
		t.Fatal("Generate returned nil error for blank user content")
	}
}

func TestChatModelStreamIsUnsupported(t *testing.T) {
	chatModel := NewChatModel("unused")
	input := []*schema.Message{schema.UserMessage("hello")}

	if _, err := chatModel.Stream(context.Background(), input); err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("Stream error = %v, want not supported error", err)
	}
}

package fake

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestStreamingChatModelStreamReturnsAssistantChunks(t *testing.T) {
	ctx := context.Background()
	chatModel := NewStreamingChatModel("hello", " stream")
	input := []*schema.Message{
		schema.SystemMessage("You are a helpful tutor."),
		schema.UserMessage("What is streaming?"),
	}

	reader, err := chatModel.Stream(ctx, input)
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	defer reader.Close()

	first, err := reader.Recv()
	if err != nil {
		t.Fatalf("first Recv returned error: %v", err)
	}
	if first.Role != schema.Assistant || first.Content != "hello" {
		t.Fatalf("first chunk = (%q, %q), want assistant hello", first.Role, first.Content)
	}

	second, err := reader.Recv()
	if err != nil {
		t.Fatalf("second Recv returned error: %v", err)
	}
	if second.Role != schema.Assistant || second.Content != " stream" {
		t.Fatalf("second chunk = (%q, %q), want assistant stream", second.Role, second.Content)
	}

	if _, err := reader.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final Recv error = %v, want io.EOF", err)
	}
	if chatModel.StreamCalls() != 1 {
		t.Fatalf("StreamCalls = %d, want 1", chatModel.StreamCalls())
	}
	if got := chatModel.LastInput(); len(got) != len(input) {
		t.Fatalf("LastInput length = %d, want %d", len(got), len(input))
	}
}

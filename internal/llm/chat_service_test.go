package llm

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
)

func TestChatServiceAskSendsUserMessageAndReturnsModelContent(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("Eino helps build LLM applications in Go.")
	service := NewChatService(chatModel)

	got, err := service.Ask(ctx, "What does Eino do?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if got != "Eino helps build LLM applications in Go." {
		t.Fatalf("answer = %q, want %q", got, "Eino helps build LLM applications in Go.")
	}

	lastInput := chatModel.LastInput()
	if len(lastInput) != 1 {
		t.Fatalf("model input length = %d, want 1", len(lastInput))
	}
	if lastInput[0].Role != schema.User {
		t.Fatalf("message role = %q, want %q", lastInput[0].Role, schema.User)
	}
	if lastInput[0].Content != "What does Eino do?" {
		t.Fatalf("message content = %q, want %q", lastInput[0].Content, "What does Eino do?")
	}
}

func TestChatServiceAskRejectsBlankQuestionBeforeCallingModel(t *testing.T) {
	chatModel := fake.NewChatModel("unused")
	service := NewChatService(chatModel)

	_, err := service.Ask(context.Background(), " \t\n ")
	if err == nil {
		t.Fatal("Ask returned nil error for blank question")
	}

	if got := chatModel.LastInput(); len(got) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(got))
	}
}

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
	if len(lastInput) != 2 {
		t.Fatalf("model input length = %d, want 2", len(lastInput))
	}
	if lastInput[0].Role != schema.System {
		t.Fatalf("first message role = %q, want %q", lastInput[0].Role, schema.System)
	}
	if lastInput[0].Content != DefaultSystemPrompt {
		t.Fatalf("first message content = %q, want %q", lastInput[0].Content, DefaultSystemPrompt)
	}
	if lastInput[1].Role != schema.User {
		t.Fatalf("second message role = %q, want %q", lastInput[1].Role, schema.User)
	}
	if lastInput[1].Content != "What does Eino do?" {
		t.Fatalf("second message content = %q, want %q", lastInput[1].Content, "What does Eino do?")
	}
}

func TestDefaultChatTemplateFormatsSystemHistoryAndQuestion(t *testing.T) {
	ctx := context.Background()
	history := []*schema.Message{
		schema.UserMessage("What is Eino?"),
		schema.AssistantMessage("Eino is a Go framework for LLM apps.", nil),
	}

	got, err := DefaultChatTemplate().Format(ctx, map[string]any{
		"history":  history,
		"question": "How does ChatTemplate help?",
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	assertMessages(t, got, []messageWant{
		{role: schema.System, content: DefaultSystemPrompt},
		{role: schema.User, content: "What is Eino?"},
		{role: schema.Assistant, content: "Eino is a Go framework for LLM apps."},
		{role: schema.User, content: "How does ChatTemplate help?"},
	})
}

func TestDefaultChatTemplateFormatsWithoutHistory(t *testing.T) {
	got, err := DefaultChatTemplate().Format(context.Background(), map[string]any{
		"question": "How does ChatTemplate help?",
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	assertMessages(t, got, []messageWant{
		{role: schema.System, content: DefaultSystemPrompt},
		{role: schema.User, content: "How does ChatTemplate help?"},
	})
}

func TestChatServiceAskWithHistoryFormatsMessagesAndReturnsModelContent(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("Templates turn variables into chat messages.")
	service := NewChatService(chatModel)
	history := []*schema.Message{
		schema.UserMessage("What did chapter 1 cover?"),
		schema.AssistantMessage("It covered fake ChatModel basics.", nil),
	}

	got, err := service.AskWithHistory(ctx, "What does chapter 2 add?", history)
	if err != nil {
		t.Fatalf("AskWithHistory returned error: %v", err)
	}

	if got != "Templates turn variables into chat messages." {
		t.Fatalf("answer = %q, want %q", got, "Templates turn variables into chat messages.")
	}

	assertMessages(t, chatModel.LastInput(), []messageWant{
		{role: schema.System, content: DefaultSystemPrompt},
		{role: schema.User, content: "What did chapter 1 cover?"},
		{role: schema.Assistant, content: "It covered fake ChatModel basics."},
		{role: schema.User, content: "What does chapter 2 add?"},
	})
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

func TestChatServiceAskWithHistoryRejectsBlankQuestionBeforeCallingModel(t *testing.T) {
	chatModel := fake.NewChatModel("unused")
	service := NewChatService(chatModel)

	_, err := service.AskWithHistory(context.Background(), " \t\n ", []*schema.Message{
		schema.UserMessage("history should not be used"),
	})
	if err == nil {
		t.Fatal("AskWithHistory returned nil error for blank question")
	}

	if got := chatModel.LastInput(); len(got) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(got))
	}
}

type messageWant struct {
	role    schema.RoleType
	content string
}

func assertMessages(t *testing.T, got []*schema.Message, want []messageWant) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("message length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Role != want[i].role {
			t.Fatalf("message[%d].Role = %q, want %q", i, got[i].Role, want[i].role)
		}
		if got[i].Content != want[i].content {
			t.Fatalf("message[%d].Content = %q, want %q", i, got[i].Content, want[i].content)
		}
	}
}

package prompting

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var ErrBlankQuestion = errors.New("chat service: question must not be blank")

const DefaultSystemPrompt = "You are a helpful Eino tutor. Explain concepts clearly and keep answers concise."

func DefaultChatTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(DefaultSystemPrompt),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{question}"),
	)
}

func FormatMessages(ctx context.Context, template prompt.ChatTemplate, question string, history []*schema.Message) ([]*schema.Message, error) {
	vars := ChatInput(question, history)

	messages, err := template.Format(ctx, vars)
	if err != nil {
		return nil, fmt.Errorf("format prompt: %w", err)
	}

	return messages, nil
}

func ChatInput(question string, history []*schema.Message) map[string]any {
	input := map[string]any{"question": question}
	if len(history) > 0 {
		input["history"] = history
	}

	return input
}

func CloneMessages(messages []*schema.Message) []*schema.Message {
	return append([]*schema.Message(nil), messages...)
}

func CloneVariables(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	copied := make(map[string]any, len(input))
	for key, value := range input {
		if messages, ok := value.([]*schema.Message); ok {
			copied[key] = CloneMessages(messages)
			continue
		}
		copied[key] = value
	}

	return copied
}

func IsBlankQuestion(question string) bool {
	return strings.TrimSpace(question) == ""
}

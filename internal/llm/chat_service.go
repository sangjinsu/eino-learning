package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var ErrBlankQuestion = errors.New("chat service: question must not be blank")

const DefaultSystemPrompt = "You are a helpful Eino tutor. Explain concepts clearly and keep answers concise."

type ChatService struct {
	model    model.BaseChatModel
	template prompt.ChatTemplate
}

func NewChatService(chatModel model.BaseChatModel) *ChatService {
	return &ChatService{
		model:    chatModel,
		template: DefaultChatTemplate(),
	}
}

func DefaultChatTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(DefaultSystemPrompt),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{question}"),
	)
}

func (s *ChatService) Ask(ctx context.Context, question string) (string, error) {
	return s.AskWithHistory(ctx, question, nil)
}

func (s *ChatService) AskWithHistory(ctx context.Context, question string, history []*schema.Message) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", ErrBlankQuestion
	}
	if s.model == nil {
		return "", errors.New("chat service: model is required")
	}
	if s.template == nil {
		return "", errors.New("chat service: template is required")
	}

	vars := map[string]any{"question": question}
	if len(history) > 0 {
		vars["history"] = history
	}

	messages, err := s.template.Format(ctx, vars)
	if err != nil {
		return "", fmt.Errorf("format prompt: %w", err)
	}

	msg, err := s.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	return msg.Content, nil
}

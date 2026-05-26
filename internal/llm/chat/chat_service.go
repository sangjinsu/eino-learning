package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

type Service struct {
	model    model.BaseChatModel
	template prompt.ChatTemplate
}

func NewService(chatModel model.BaseChatModel) *Service {
	return NewServiceWithTemplate(chatModel, prompting.DefaultChatTemplate())
}

func NewServiceWithTemplate(chatModel model.BaseChatModel, template prompt.ChatTemplate) *Service {
	return &Service{
		model:    chatModel,
		template: template,
	}
}

func (s *Service) Ask(ctx context.Context, question string) (string, error) {
	return s.AskWithHistory(ctx, question, nil)
}

func (s *Service) AskWithHistory(ctx context.Context, question string, history []*schema.Message) (string, error) {
	if prompting.IsBlankQuestion(question) {
		return "", prompting.ErrBlankQuestion
	}
	if s.model == nil {
		return "", errors.New("chat service: model is required")
	}
	if s.template == nil {
		return "", errors.New("chat service: template is required")
	}

	messages, err := prompting.FormatMessages(ctx, s.template, question, history)
	if err != nil {
		return "", err
	}

	msg, err := s.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	return msg.Content, nil
}

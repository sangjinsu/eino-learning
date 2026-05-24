package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

var ErrBlankQuestion = errors.New("chat service: question must not be blank")

type ChatService struct {
	model model.BaseChatModel
}

func NewChatService(chatModel model.BaseChatModel) *ChatService {
	return &ChatService{model: chatModel}
}

func (s *ChatService) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", ErrBlankQuestion
	}
	if s.model == nil {
		return "", errors.New("chat service: model is required")
	}

	msg, err := s.model.Generate(ctx, []*schema.Message{
		schema.UserMessage(question),
	})
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	return msg.Content, nil
}

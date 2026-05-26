package openai

import (
	"context"
	"fmt"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func NewChatModel(ctx context.Context, cfg Config) (model.BaseChatModel, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	chatModel, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai chat model: %w", err)
	}

	return chatModel, nil
}

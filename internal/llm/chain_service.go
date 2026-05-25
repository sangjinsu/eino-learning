package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

var (
	ErrChainModelRequired    = errors.New("chat chain: model is required")
	ErrChainTemplateRequired = errors.New("chat chain: template is required")
)

type ChatChainService struct {
	runnable compose.Runnable[map[string]any, *schema.Message]
}

type ChatChainTrace struct {
	InputVariables map[string]any
	PromptMessages []*schema.Message
	ModelResponse  *schema.Message
}

func (t *ChatChainTrace) Answer() string {
	if t == nil || t.ModelResponse == nil {
		return ""
	}

	return t.ModelResponse.Content
}

func NewChatChainService(ctx context.Context, chatModel model.BaseChatModel) (*ChatChainService, error) {
	return NewChatChainServiceWithTemplate(ctx, chatModel, DefaultChatTemplate())
}

func NewChatChainServiceWithTemplate(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate) (*ChatChainService, error) {
	runnable, err := NewChatChain(ctx, chatModel, template)
	if err != nil {
		return nil, err
	}

	return &ChatChainService{runnable: runnable}, nil
}

func NewChatChain(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate) (compose.Runnable[map[string]any, *schema.Message], error) {
	if chatModel == nil {
		return nil, ErrChainModelRequired
	}
	if template == nil {
		return nil, ErrChainTemplateRequired
	}

	chain := compose.NewChain[map[string]any, *schema.Message]().
		AppendChatTemplate(template).
		AppendChatModel(chatModel)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile chat chain: %w", err)
	}

	return runnable, nil
}

func NewTracedChatChain(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, trace *ChatChainTrace) (compose.Runnable[map[string]any, *schema.Message], error) {
	if chatModel == nil {
		return nil, ErrChainModelRequired
	}
	if template == nil {
		return nil, ErrChainTemplateRequired
	}
	if trace == nil {
		trace = &ChatChainTrace{}
	}

	chain := compose.NewChain[map[string]any, *schema.Message]().
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			_ = ctx
			trace.InputVariables = cloneVariables(input)
			return input, nil
		})).
		AppendChatTemplate(template).
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, messages []*schema.Message) ([]*schema.Message, error) {
			_ = ctx
			trace.PromptMessages = cloneMessages(messages)
			return messages, nil
		})).
		AppendChatModel(chatModel).
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, message *schema.Message) (*schema.Message, error) {
			_ = ctx
			trace.ModelResponse = message
			return message, nil
		}))

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile traced chat chain: %w", err)
	}

	return runnable, nil
}

func RunChatChainWithTrace(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, question string, history []*schema.Message) (*ChatChainTrace, error) {
	if strings.TrimSpace(question) == "" {
		return nil, ErrBlankQuestion
	}

	trace := &ChatChainTrace{}
	runnable, err := NewTracedChatChain(ctx, chatModel, template, trace)
	if err != nil {
		return nil, err
	}

	message, err := runnable.Invoke(ctx, chatChainInput(question, history))
	if err != nil {
		return nil, fmt.Errorf("invoke traced chat chain: %w", err)
	}
	if message == nil {
		return nil, errors.New("invoke traced chat chain: empty response")
	}
	if trace.ModelResponse == nil {
		trace.ModelResponse = message
	}

	return trace, nil
}

func (s *ChatChainService) Ask(ctx context.Context, question string) (string, error) {
	return s.AskWithHistory(ctx, question, nil)
}

func (s *ChatChainService) AskWithHistory(ctx context.Context, question string, history []*schema.Message) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", ErrBlankQuestion
	}
	if s == nil || s.runnable == nil {
		return "", errors.New("chat chain: runnable is required")
	}

	msg, err := s.runnable.Invoke(ctx, chatChainInput(question, history))
	if err != nil {
		return "", fmt.Errorf("invoke chat chain: %w", err)
	}
	if msg == nil {
		return "", errors.New("invoke chat chain: empty response")
	}

	return msg.Content, nil
}

func chatChainInput(question string, history []*schema.Message) map[string]any {
	input := map[string]any{"question": question}
	if len(history) > 0 {
		input["history"] = history
	}

	return input
}

func cloneVariables(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	copied := make(map[string]any, len(input))
	for key, value := range input {
		if messages, ok := value.([]*schema.Message); ok {
			copied[key] = cloneMessages(messages)
			continue
		}
		copied[key] = value
	}

	return copied
}

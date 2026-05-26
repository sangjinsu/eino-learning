package chain

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

var (
	ErrChainModelRequired    = errors.New("chat chain: model is required")
	ErrChainTemplateRequired = errors.New("chat chain: template is required")
)

type Service struct {
	runnable compose.Runnable[map[string]any, *schema.Message]
}

type Trace struct {
	InputVariables map[string]any
	PromptMessages []*schema.Message
	ModelResponse  *schema.Message
}

func (t *Trace) Answer() string {
	if t == nil || t.ModelResponse == nil {
		return ""
	}

	return t.ModelResponse.Content
}

func NewService(ctx context.Context, chatModel model.BaseChatModel) (*Service, error) {
	return NewServiceWithTemplate(ctx, chatModel, prompting.DefaultChatTemplate())
}

func NewServiceWithTemplate(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate) (*Service, error) {
	runnable, err := NewRunnable(ctx, chatModel, template)
	if err != nil {
		return nil, err
	}

	return &Service{runnable: runnable}, nil
}

func NewRunnable(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate) (compose.Runnable[map[string]any, *schema.Message], error) {
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

func NewTracedRunnable(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, trace *Trace) (compose.Runnable[map[string]any, *schema.Message], error) {
	if chatModel == nil {
		return nil, ErrChainModelRequired
	}
	if template == nil {
		return nil, ErrChainTemplateRequired
	}
	if trace == nil {
		trace = &Trace{}
	}

	chain := compose.NewChain[map[string]any, *schema.Message]().
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			_ = ctx
			trace.InputVariables = prompting.CloneVariables(input)
			return input, nil
		})).
		AppendChatTemplate(template).
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, messages []*schema.Message) ([]*schema.Message, error) {
			_ = ctx
			trace.PromptMessages = prompting.CloneMessages(messages)
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

func RunWithTrace(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, question string, history []*schema.Message) (*Trace, error) {
	if prompting.IsBlankQuestion(question) {
		return nil, prompting.ErrBlankQuestion
	}

	trace := &Trace{}
	runnable, err := NewTracedRunnable(ctx, chatModel, template, trace)
	if err != nil {
		return nil, err
	}

	message, err := runnable.Invoke(ctx, prompting.ChatInput(question, history))
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

func (s *Service) Ask(ctx context.Context, question string) (string, error) {
	return s.AskWithHistory(ctx, question, nil)
}

func (s *Service) AskWithHistory(ctx context.Context, question string, history []*schema.Message) (string, error) {
	if prompting.IsBlankQuestion(question) {
		return "", prompting.ErrBlankQuestion
	}
	if s == nil || s.runnable == nil {
		return "", errors.New("chat chain: runnable is required")
	}

	msg, err := s.runnable.Invoke(ctx, prompting.ChatInput(question, history))
	if err != nil {
		return "", fmt.Errorf("invoke chat chain: %w", err)
	}
	if msg == nil {
		return "", errors.New("invoke chat chain: empty response")
	}

	return msg.Content, nil
}

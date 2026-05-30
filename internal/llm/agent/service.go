package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	react "github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

const DefaultMaxStep = 12

var (
	ErrModelRequired = errors.New("react agent service: tool calling model is required")
	ErrToolRequired  = errors.New("react agent service: tool must not be nil")
)

type Options struct {
	MaxStep int
}

type Service struct {
	model   model.ToolCallingChatModel
	tools   []einotool.BaseTool
	options Options
}

type Result struct {
	Question       string
	FinalResponse  *schema.Message
	AvailableTools []string
	MaxStep        int
}

func NewService(chatModel model.ToolCallingChatModel, tools []einotool.BaseTool) *Service {
	return NewServiceWithOptions(chatModel, tools, Options{})
}

func NewServiceWithOptions(chatModel model.ToolCallingChatModel, tools []einotool.BaseTool, options Options) *Service {
	if options.MaxStep <= 0 {
		options.MaxStep = DefaultMaxStep
	}

	return &Service{
		model:   chatModel,
		tools:   append([]einotool.BaseTool(nil), tools...),
		options: options,
	}
}

func (r *Result) Answer() string {
	if r == nil || r.FinalResponse == nil {
		return ""
	}

	return r.FinalResponse.Content
}

func (s *Service) Ask(ctx context.Context, question string) (*Result, error) {
	if prompting.IsBlankQuestion(question) {
		return nil, prompting.ErrBlankQuestion
	}
	if s.model == nil {
		return nil, ErrModelRequired
	}

	toolNames, err := toolNames(ctx, s.tools)
	if err != nil {
		return nil, err
	}

	reactAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: s.model,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools:               s.tools,
			ExecuteSequentially: true,
		},
		MaxStep: s.options.MaxStep,
	})
	if err != nil {
		return nil, fmt.Errorf("create react agent: %w", err)
	}

	finalResponse, err := reactAgent.Generate(ctx, []*schema.Message{
		schema.UserMessage(question),
	})
	if err != nil {
		return nil, fmt.Errorf("run react agent: %w", err)
	}
	if finalResponse == nil {
		return nil, errors.New("run react agent: empty response")
	}

	return &Result{
		Question:       question,
		FinalResponse:  finalResponse,
		AvailableTools: toolNames,
		MaxStep:        s.options.MaxStep,
	}, nil
}

func toolNames(ctx context.Context, tools []einotool.BaseTool) ([]string, error) {
	names := make([]string, 0, len(tools))
	for _, candidate := range tools {
		if candidate == nil {
			return nil, ErrToolRequired
		}

		info, err := candidate.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("read tool info: %w", err)
		}
		names = append(names, info.Name)
	}

	return names, nil
}

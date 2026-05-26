package toolcalling

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
	toolrunner "github.com/sangjinsu/eino-learning/internal/tools"
)

var ErrToolCallingModelRequired = errors.New("chat service: model must support tool calling")

type Service struct {
	model    model.BaseChatModel
	template prompt.ChatTemplate
}

type Result struct {
	PromptMessages []*schema.Message
	FirstResponse  *schema.Message
	ToolMessages   []*schema.Message
	FinalResponse  *schema.Message
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

func (r *Result) Answer() string {
	if r == nil || r.FinalResponse == nil {
		return ""
	}

	return r.FinalResponse.Content
}

func (s *Service) Ask(ctx context.Context, question string, allowedTools []einotool.BaseTool) (*Result, error) {
	return s.AskWithHistoryAndTools(ctx, question, nil, allowedTools)
}

func (s *Service) AskWithHistoryAndTools(ctx context.Context, question string, history []*schema.Message, allowedTools []einotool.BaseTool) (*Result, error) {
	if prompting.IsBlankQuestion(question) {
		return nil, prompting.ErrBlankQuestion
	}
	if s.model == nil {
		return nil, errors.New("chat service: model is required")
	}
	if s.template == nil {
		return nil, errors.New("chat service: template is required")
	}

	toolCallingModel, ok := s.model.(model.ToolCallingChatModel)
	if !ok {
		return nil, ErrToolCallingModelRequired
	}

	messages, err := prompting.FormatMessages(ctx, s.template, question, history)
	if err != nil {
		return nil, err
	}

	toolInfos, err := collectToolInfos(ctx, allowedTools)
	if err != nil {
		return nil, err
	}
	modelWithTools, err := toolCallingModel.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("bind tools: %w", err)
	}

	firstResponse, err := modelWithTools.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate tool call: %w", err)
	}
	if firstResponse == nil {
		return nil, errors.New("generate tool call: empty response")
	}
	if len(firstResponse.ToolCalls) == 0 {
		return &Result{
			PromptMessages: prompting.CloneMessages(messages),
			FirstResponse:  firstResponse,
			FinalResponse:  firstResponse,
		}, nil
	}

	toolMessages, err := toolrunner.ExecuteToolCalls(ctx, allowedTools, firstResponse)
	if err != nil {
		return nil, fmt.Errorf("execute tool calls: %w", err)
	}

	followupMessages := append(prompting.CloneMessages(messages), firstResponse)
	followupMessages = append(followupMessages, toolMessages...)

	finalResponse, err := modelWithTools.Generate(ctx, followupMessages)
	if err != nil {
		return nil, fmt.Errorf("generate final answer: %w", err)
	}
	if finalResponse == nil {
		return nil, errors.New("generate final answer: empty response")
	}

	return &Result{
		PromptMessages: prompting.CloneMessages(messages),
		FirstResponse:  firstResponse,
		ToolMessages:   prompting.CloneMessages(toolMessages),
		FinalResponse:  finalResponse,
	}, nil
}

func collectToolInfos(ctx context.Context, allowedTools []einotool.BaseTool) ([]*schema.ToolInfo, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(allowedTools))
	for _, allowedTool := range allowedTools {
		if allowedTool == nil {
			return nil, errors.New("chat service: tool must not be nil")
		}

		info, err := allowedTool.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("read tool info: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}

	return toolInfos, nil
}

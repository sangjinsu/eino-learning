package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	toolrunner "github.com/sangjinsu/eino-learning/internal/tools"
)

var ErrToolCallingModelRequired = errors.New("chat service: model must support tool calling")

type ToolCallingResult struct {
	PromptMessages []*schema.Message
	FirstResponse  *schema.Message
	ToolMessages   []*schema.Message
	FinalResponse  *schema.Message
}

func (r *ToolCallingResult) Answer() string {
	if r == nil || r.FinalResponse == nil {
		return ""
	}

	return r.FinalResponse.Content
}

func (s *ChatService) AskWithTools(ctx context.Context, question string, allowedTools []einotool.BaseTool) (*ToolCallingResult, error) {
	return s.AskWithHistoryAndTools(ctx, question, nil, allowedTools)
}

func (s *ChatService) AskWithHistoryAndTools(ctx context.Context, question string, history []*schema.Message, allowedTools []einotool.BaseTool) (*ToolCallingResult, error) {
	if strings.TrimSpace(question) == "" {
		return nil, ErrBlankQuestion
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

	messages, err := s.formatMessages(ctx, question, history)
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
		return &ToolCallingResult{
			PromptMessages: cloneMessages(messages),
			FirstResponse:  firstResponse,
			FinalResponse:  firstResponse,
		}, nil
	}

	toolMessages, err := toolrunner.ExecuteToolCalls(ctx, allowedTools, firstResponse)
	if err != nil {
		return nil, fmt.Errorf("execute tool calls: %w", err)
	}

	followupMessages := append(cloneMessages(messages), firstResponse)
	followupMessages = append(followupMessages, toolMessages...)

	finalResponse, err := modelWithTools.Generate(ctx, followupMessages)
	if err != nil {
		return nil, fmt.Errorf("generate final answer: %w", err)
	}
	if finalResponse == nil {
		return nil, errors.New("generate final answer: empty response")
	}

	return &ToolCallingResult{
		PromptMessages: cloneMessages(messages),
		FirstResponse:  firstResponse,
		ToolMessages:   cloneMessages(toolMessages),
		FinalResponse:  finalResponse,
	}, nil
}

func (s *ChatService) formatMessages(ctx context.Context, question string, history []*schema.Message) ([]*schema.Message, error) {
	vars := map[string]any{"question": question}
	if len(history) > 0 {
		vars["history"] = history
	}

	messages, err := s.template.Format(ctx, vars)
	if err != nil {
		return nil, fmt.Errorf("format prompt: %w", err)
	}

	return messages, nil
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

func cloneMessages(messages []*schema.Message) []*schema.Message {
	return append([]*schema.Message(nil), messages...)
}

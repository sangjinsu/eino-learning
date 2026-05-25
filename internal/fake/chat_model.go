package fake

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

var (
	ErrEmptyInput     = errors.New("fake chat model: input messages are required")
	ErrBlankUserInput = errors.New("fake chat model: last user message must not be blank")
	ErrNotSupported   = errors.New("fake chat model: operation not supported in chapter 01")
)

type ChatModel struct {
	response  string
	lastInput []*schema.Message
}

var _ model.BaseChatModel = (*ChatModel)(nil)

func NewChatModel(response string) *ChatModel {
	return &ChatModel{response: response}
}

func (m *ChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	_ = ctx
	_ = opts

	if err := validateChatInput(input); err != nil {
		return nil, err
	}

	m.lastInput = append([]*schema.Message(nil), input...)
	return schema.AssistantMessage(m.response, nil), nil
}

func (m *ChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	_ = ctx
	_ = input
	_ = opts

	return nil, ErrNotSupported
}

func (m *ChatModel) LastInput() []*schema.Message {
	return append([]*schema.Message(nil), m.lastInput...)
}

func validateChatInput(input []*schema.Message) error {
	if len(input) == 0 {
		return ErrEmptyInput
	}

	last := input[len(input)-1]
	if last == nil || last.Role != schema.User || strings.TrimSpace(last.Content) == "" {
		return ErrBlankUserInput
	}

	return nil
}

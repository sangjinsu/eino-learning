package fake

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type StreamingChatModel struct {
	chunks      []string
	lastInput   []*schema.Message
	streamCalls int
}

var _ model.BaseChatModel = (*StreamingChatModel)(nil)

func NewStreamingChatModel(chunks ...string) *StreamingChatModel {
	return &StreamingChatModel{chunks: append([]string(nil), chunks...)}
}

func (m *StreamingChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	_ = ctx
	_ = opts

	if err := validateChatInput(input); err != nil {
		return nil, err
	}

	m.lastInput = append([]*schema.Message(nil), input...)
	return schema.AssistantMessage(strings.Join(m.chunks, ""), nil), nil
}

func (m *StreamingChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	_ = ctx
	_ = opts

	if err := validateChatInput(input); err != nil {
		return nil, err
	}

	m.lastInput = append([]*schema.Message(nil), input...)
	m.streamCalls++

	chunks := make([]*schema.Message, 0, len(m.chunks))
	for _, chunk := range m.chunks {
		chunks = append(chunks, schema.AssistantMessage(chunk, nil))
	}

	return schema.StreamReaderFromArray(chunks), nil
}

func (m *StreamingChatModel) LastInput() []*schema.Message {
	return append([]*schema.Message(nil), m.lastInput...)
}

func (m *StreamingChatModel) StreamCalls() int {
	return m.streamCalls
}

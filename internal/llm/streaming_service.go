package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type StreamingResult struct {
	PromptMessages []*schema.Message
	Chunks         []*schema.Message
	Answer         string
}

// Stream은 질문을 기본 template으로 변환한 뒤 model stream을 시작합니다.
func (s *ChatService) Stream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error) {
	return s.StreamWithHistory(ctx, question, nil)
}

// StreamWithHistory는 호출자가 chunk를 직접 읽을 수 있도록 원본 StreamReader를 반환합니다.
func (s *ChatService) StreamWithHistory(ctx context.Context, question string, history []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	_, reader, err := s.openStreamWithHistory(ctx, question, history)
	return reader, err
}

// AskStreaming은 stream을 끝까지 읽어 chunk 목록과 최종 answer를 반환합니다.
func (s *ChatService) AskStreaming(ctx context.Context, question string) (*StreamingResult, error) {
	return s.AskStreamingWithHistory(ctx, question, nil)
}

// AskStreamingWithHistory는 template이 적용된 chat 요청을 streaming으로 실행하고 모든 chunk를 모읍니다.
func (s *ChatService) AskStreamingWithHistory(ctx context.Context, question string, history []*schema.Message) (*StreamingResult, error) {
	messages, reader, err := s.openStreamWithHistory(ctx, question, history)
	if err != nil {
		return nil, err
	}

	result, err := CollectMessageStream(reader)
	if err != nil {
		return nil, err
	}

	result.PromptMessages = cloneMessages(messages)
	return result, nil
}

// CollectMessageStream은 reader를 읽고 닫은 뒤 chunk content를 하나의 answer로 이어 붙입니다.
func CollectMessageStream(reader *schema.StreamReader[*schema.Message]) (*StreamingResult, error) {
	if reader == nil {
		return nil, errors.New("chat service: stream reader is required")
	}
	defer reader.Close()

	var builder strings.Builder
	chunks := []*schema.Message{}

	for {
		chunk, err := reader.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("receive stream chunk: %w", err)
		}
		if chunk == nil {
			continue
		}

		chunks = append(chunks, chunk)
		builder.WriteString(chunk.Content)
	}

	return &StreamingResult{
		Chunks: cloneMessages(chunks),
		Answer: builder.String(),
	}, nil
}

func (s *ChatService) openStreamWithHistory(ctx context.Context, question string, history []*schema.Message) ([]*schema.Message, *schema.StreamReader[*schema.Message], error) {
	if strings.TrimSpace(question) == "" {
		return nil, nil, ErrBlankQuestion
	}
	if s.model == nil {
		return nil, nil, errors.New("chat service: model is required")
	}
	if s.template == nil {
		return nil, nil, errors.New("chat service: template is required")
	}

	messages, err := s.formatMessages(ctx, question, history)
	if err != nil {
		return nil, nil, err
	}

	reader, err := s.model.Stream(ctx, messages)
	if err != nil {
		return nil, nil, fmt.Errorf("stream answer: %w", err)
	}

	return messages, reader, nil
}

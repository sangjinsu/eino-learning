package streaming

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

type Service struct {
	model    model.BaseChatModel
	template prompt.ChatTemplate
}

type Result struct {
	PromptMessages []*schema.Message
	Chunks         []*schema.Message
	Answer         string
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

// Stream은 질문을 기본 template으로 변환한 뒤 model stream을 시작합니다.
func (s *Service) Stream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error) {
	return s.StreamWithHistory(ctx, question, nil)
}

// StreamWithHistory는 호출자가 chunk를 직접 읽을 수 있도록 원본 StreamReader를 반환합니다.
func (s *Service) StreamWithHistory(ctx context.Context, question string, history []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	_, reader, err := s.openStreamWithHistory(ctx, question, history)
	return reader, err
}

// Ask은 stream을 끝까지 읽어 chunk 목록과 최종 answer를 반환합니다.
func (s *Service) Ask(ctx context.Context, question string) (*Result, error) {
	return s.AskWithHistory(ctx, question, nil)
}

// AskWithHistory는 template이 적용된 chat 요청을 streaming으로 실행하고 모든 chunk를 모읍니다.
func (s *Service) AskWithHistory(ctx context.Context, question string, history []*schema.Message) (*Result, error) {
	messages, reader, err := s.openStreamWithHistory(ctx, question, history)
	if err != nil {
		return nil, err
	}

	result, err := CollectMessageStream(reader)
	if err != nil {
		return nil, err
	}

	result.PromptMessages = prompting.CloneMessages(messages)
	return result, nil
}

// CollectMessageStream은 reader를 읽고 닫은 뒤 chunk content를 하나의 answer로 이어 붙입니다.
func CollectMessageStream(reader *schema.StreamReader[*schema.Message]) (*Result, error) {
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

	return &Result{
		Chunks: prompting.CloneMessages(chunks),
		Answer: builder.String(),
	}, nil
}

func (s *Service) openStreamWithHistory(ctx context.Context, question string, history []*schema.Message) ([]*schema.Message, *schema.StreamReader[*schema.Message], error) {
	if prompting.IsBlankQuestion(question) {
		return nil, nil, prompting.ErrBlankQuestion
	}
	if s.model == nil {
		return nil, nil, errors.New("chat service: model is required")
	}
	if s.template == nil {
		return nil, nil, errors.New("chat service: template is required")
	}

	messages, err := prompting.FormatMessages(ctx, s.template, question, history)
	if err != nil {
		return nil, nil, err
	}

	reader, err := s.model.Stream(ctx, messages)
	if err != nil {
		return nil, nil, fmt.Errorf("stream answer: %w", err)
	}

	return messages, reader, nil
}

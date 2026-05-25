package llm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	observableNodePrompt = "prompt"
	observableNodeModel  = "model"
)

type CallbackTiming string

const (
	CallbackTimingStart CallbackTiming = "start"
	CallbackTimingEnd   CallbackTiming = "end"
	CallbackTimingError CallbackTiming = "error"
)

// CallbackEvent는 Eino callback 한 번을 학습용으로 읽기 쉽게 요약한 값입니다.
type CallbackEvent struct {
	Timing    CallbackTiming
	Name      string
	Component string
	Summary   string
}

// ObservableChatResult는 model 답변과 실행 중 수집된 callback event를 함께 담습니다.
type ObservableChatResult struct {
	Answer string
	Events []CallbackEvent
}

// CallbackRecorder는 callback handler에서 호출되는 event를 thread-safe하게 모읍니다.
type CallbackRecorder struct {
	mu     sync.Mutex
	events []CallbackEvent
}

// NewCallbackRecorder는 한 번의 실행을 관찰할 recorder를 생성합니다.
func NewCallbackRecorder() *CallbackRecorder {
	return &CallbackRecorder{}
}

// Handler는 Eino component lifecycle callback을 CallbackEvent로 기록합니다.
func (r *CallbackRecorder) Handler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			r.record(CallbackEvent{
				Timing:    CallbackTimingStart,
				Name:      callbackName(info),
				Component: callbackComponent(info),
				Summary:   summarizeCallbackInput(input),
			})
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			r.record(CallbackEvent{
				Timing:    CallbackTimingEnd,
				Name:      callbackName(info),
				Component: callbackComponent(info),
				Summary:   summarizeCallbackOutput(output),
			})
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			r.record(CallbackEvent{
				Timing:    CallbackTimingError,
				Name:      callbackName(info),
				Component: callbackComponent(info),
				Summary:   err.Error(),
			})
			return ctx
		}).
		Build()
}

// Events는 지금까지 기록된 callback event snapshot을 반환합니다.
func (r *CallbackRecorder) Events() []CallbackEvent {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]CallbackEvent(nil), r.events...)
}

func (r *CallbackRecorder) record(event CallbackEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
}

// RunObservableChatChain은 ChatTemplate -> ChatModel Chain을 실행하면서 callback event를 수집합니다.
func RunObservableChatChain(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, question string, history []*schema.Message) (*ObservableChatResult, error) {
	if strings.TrimSpace(question) == "" {
		return nil, ErrBlankQuestion
	}

	runnable, err := NewObservableChatChain(ctx, chatModel, template)
	if err != nil {
		return nil, err
	}

	recorder := NewCallbackRecorder()
	message, err := runnable.Invoke(ctx, chatChainInput(question, history), compose.WithCallbacks(recorder.Handler()))
	result := &ObservableChatResult{Events: recorder.Events()}
	if err != nil {
		return result, fmt.Errorf("invoke observable chat chain: %w", err)
	}
	if message == nil {
		return result, errors.New("invoke observable chat chain: empty response")
	}

	result.Answer = message.Content
	return result, nil
}

// NewObservableChatChain은 callback 출력에서 알아보기 쉽도록 node 이름을 지정한 Chain을 만듭니다.
func NewObservableChatChain(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate) (compose.Runnable[map[string]any, *schema.Message], error) {
	if chatModel == nil {
		return nil, ErrChainModelRequired
	}
	if template == nil {
		return nil, ErrChainTemplateRequired
	}

	chain := compose.NewChain[map[string]any, *schema.Message]().
		AppendChatTemplate(
			template,
			compose.WithNodeKey(observableNodePrompt),
			compose.WithNodeName("ChatTemplate"),
		).
		AppendChatModel(
			chatModel,
			compose.WithNodeKey(observableNodeModel),
			compose.WithNodeName("ChatModel"),
		)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile observable chat chain: %w", err)
	}

	return runnable, nil
}

// callbackName은 RunInfo가 없을 수 있는 상황을 안전하게 처리합니다.
func callbackName(info *callbacks.RunInfo) string {
	if info == nil {
		return ""
	}

	return info.Name
}

// callbackComponent는 component 종류를 log label로 쓰기 쉬운 문자열로 바꿉니다.
func callbackComponent(info *callbacks.RunInfo) string {
	if info == nil {
		return ""
	}

	return fmt.Sprint(info.Component)
}

// summarizeCallbackInput은 component별 callback input을 짧은 학습용 문장으로 요약합니다.
func summarizeCallbackInput(input callbacks.CallbackInput) string {
	if promptInput := prompt.ConvCallbackInput(input); promptInput != nil {
		return fmt.Sprintf("variables=%s", sortedVariableKeys(promptInput.Variables))
	}

	if modelInput := model.ConvCallbackInput(input); modelInput != nil {
		return fmt.Sprintf("messages=%d", len(modelInput.Messages))
	}

	return fmt.Sprintf("input=%T", input)
}

// summarizeCallbackOutput은 component별 callback output을 console에서 읽기 좋게 요약합니다.
func summarizeCallbackOutput(output callbacks.CallbackOutput) string {
	if promptOutput := prompt.ConvCallbackOutput(output); promptOutput != nil {
		return fmt.Sprintf("messages=%d", len(promptOutput.Result))
	}

	if modelOutput := model.ConvCallbackOutput(output); modelOutput != nil && modelOutput.Message != nil {
		return fmt.Sprintf("role=%s content=%s", modelOutput.Message.Role, summarizeContent(modelOutput.Message.Content))
	}

	return fmt.Sprintf("output=%T", output)
}

// summarizeContent는 긴 model 응답을 callback summary에 맞게 한 줄로 줄입니다.
func summarizeContent(content string) string {
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) <= 160 {
		return content
	}

	return string(runes[:157]) + "..."
}

// sortedVariableKeys는 callback 출력이 매번 같은 순서로 보이도록 변수 이름을 정렬합니다.
func sortedVariableKeys(variables map[string]any) string {
	if len(variables) == 0 {
		return "[]"
	}

	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return "[" + strings.Join(keys, ",") + "]"
}

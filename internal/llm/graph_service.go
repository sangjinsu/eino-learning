package llm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

const (
	GraphRouteChat       = "chat"
	GraphRouteCalculator = "calculator"
)

// 노드 이름을 상수로 고정해 CLI 출력, 테스트, 문서가 같은 그래프를 설명하게 합니다.
const (
	graphNodeRoute       = "route"
	graphNodeCalculator  = "calculator"
	graphNodePrepare     = "prepare_prompt"
	graphNodePrompt      = "prompt"
	graphNodeTracePrompt = "trace_prompt"
	graphNodeModel       = "model"
	graphNodeModelOutput = "model_output"
)

var (
	ErrGraphModelRequired    = errors.New("assistant graph: model is required")
	ErrGraphTemplateRequired = errors.New("assistant graph: template is required")
)

type AssistantGraphInput struct {
	Question string
	History  []*schema.Message
}

// AssistantGraphResult는 두 graph branch가 공유하는 단일 출력 타입입니다.
// 선택된 route에서 생성한 필드만 채워집니다.
type AssistantGraphResult struct {
	Route          string
	Answer         string
	Calculation    *tools.CalculatorOutput
	PromptMessages []*schema.Message
	ModelResponse  *schema.Message
}

type AssistantGraphService struct {
	model    model.BaseChatModel
	template prompt.ChatTemplate
}

// assistantGraphState는 route node에서 선택된 branch로 넘기는 내부 값입니다.
// 공개 입력값과 routing metadata를 함께 들고 갑니다.
type assistantGraphState struct {
	Question   string
	History    []*schema.Message
	Route      string
	Expression string
}

func NewAssistantGraphService(ctx context.Context, chatModel model.BaseChatModel) (*AssistantGraphService, error) {
	return NewAssistantGraphServiceWithTemplate(ctx, chatModel, DefaultChatTemplate())
}

func NewAssistantGraphServiceWithTemplate(_ context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate) (*AssistantGraphService, error) {
	if chatModel == nil {
		return nil, ErrGraphModelRequired
	}
	if template == nil {
		return nil, ErrGraphTemplateRequired
	}

	return &AssistantGraphService{
		model:    chatModel,
		template: template,
	}, nil
}

func (s *AssistantGraphService) Run(ctx context.Context, input AssistantGraphInput) (*AssistantGraphResult, error) {
	if strings.TrimSpace(input.Question) == "" {
		return nil, ErrBlankQuestion
	}
	if s == nil || s.model == nil {
		return nil, ErrGraphModelRequired
	}
	if s.template == nil {
		return nil, ErrGraphTemplateRequired
	}

	return RunAssistantGraph(ctx, s.model, s.template, input)
}

func RunAssistantGraph(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, input AssistantGraphInput) (*AssistantGraphResult, error) {
	if strings.TrimSpace(input.Question) == "" {
		return nil, ErrBlankQuestion
	}

	trace := &AssistantGraphResult{}
	runnable, err := NewAssistantGraph(ctx, chatModel, template, trace)
	if err != nil {
		return nil, err
	}

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("invoke assistant graph: %w", err)
	}
	if result == nil {
		return nil, errors.New("invoke assistant graph: empty response")
	}

	return result, nil
}

func NewAssistantGraph(ctx context.Context, chatModel model.BaseChatModel, template prompt.ChatTemplate, trace *AssistantGraphResult) (compose.Runnable[AssistantGraphInput, *AssistantGraphResult], error) {
	if chatModel == nil {
		return nil, ErrGraphModelRequired
	}
	if template == nil {
		return nil, ErrGraphTemplateRequired
	}
	if trace == nil {
		trace = &AssistantGraphResult{}
	}

	// 그래프 모양:
	// START -> route
	// route -> calculator -> END
	// route -> prepare_prompt -> prompt -> trace_prompt -> model -> model_output -> END
	graph := compose.NewGraph[AssistantGraphInput, *AssistantGraphResult]()

	// route node는 입력을 정리하고 실행할 branch를 결정합니다.
	if err := graph.AddLambdaNode(graphNodeRoute, compose.InvokableLambda(routeAssistantGraph)); err != nil {
		return nil, err
	}
	// calculator branch는 deterministic local 작업만 수행하며 의도적으로 model을 호출하지 않습니다.
	if err := graph.AddLambdaNode(graphNodeCalculator, compose.InvokableLambda(func(ctx context.Context, state *assistantGraphState) (*AssistantGraphResult, error) {
		return runCalculatorGraphNode(ctx, state, trace)
	})); err != nil {
		return nil, err
	}
	// chat branch는 routing state를 ChatTemplate 입력 변수로 다시 변환합니다.
	if err := graph.AddLambdaNode(graphNodePrepare, compose.InvokableLambda(prepareGraphPromptVariables)); err != nil {
		return nil, err
	}
	if err := graph.AddChatTemplateNode(graphNodePrompt, template); err != nil {
		return nil, err
	}
	// trace node는 테스트와 CLI 예제에서 ChatTemplate 출력을 관찰할 수 있게 합니다.
	if err := graph.AddLambdaNode(graphNodeTracePrompt, compose.InvokableLambda(func(ctx context.Context, messages []*schema.Message) ([]*schema.Message, error) {
		_ = ctx
		trace.PromptMessages = cloneMessages(messages)
		return messages, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddChatModelNode(graphNodeModel, chatModel); err != nil {
		return nil, err
	}
	// 마지막 chat node는 model message를 graph의 공통 결과 타입으로 감쌉니다.
	if err := graph.AddLambdaNode(graphNodeModelOutput, compose.InvokableLambda(func(ctx context.Context, message *schema.Message) (*AssistantGraphResult, error) {
		_ = ctx
		trace.Route = GraphRouteChat
		trace.ModelResponse = message
		if message != nil {
			trace.Answer = message.Content
		}
		return trace, nil
	})); err != nil {
		return nil, err
	}

	if err := graph.AddEdge(compose.START, graphNodeRoute); err != nil {
		return nil, err
	}
	// Chapter 6의 핵심 학습 포인트입니다. runtime input이 다음 node를 선택합니다.
	if err := graph.AddBranch(graphNodeRoute, compose.NewGraphBranch(func(ctx context.Context, state *assistantGraphState) (string, error) {
		_ = ctx
		if state.Route == GraphRouteCalculator {
			return graphNodeCalculator, nil
		}
		return graphNodePrepare, nil
	}, map[string]bool{
		graphNodeCalculator: true,
		graphNodePrepare:    true,
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(graphNodeCalculator, compose.END); err != nil {
		return nil, err
	}
	// chat branch의 edge를 명시해 각 단계의 데이터 형태가 보이게 합니다.
	if err := graph.AddEdge(graphNodePrepare, graphNodePrompt); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(graphNodePrompt, graphNodeTracePrompt); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(graphNodeTracePrompt, graphNodeModel); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(graphNodeModel, graphNodeModelOutput); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(graphNodeModelOutput, compose.END); err != nil {
		return nil, err
	}

	runnable, err := graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile assistant graph: %w", err)
	}

	return runnable, nil
}

func routeAssistantGraph(_ context.Context, input AssistantGraphInput) (*assistantGraphState, error) {
	question := strings.TrimSpace(input.Question)
	if question == "" {
		return nil, ErrBlankQuestion
	}

	// 기본값은 model-backed chat route입니다. 명확히 계산 가능한 질문일 때만 계산 route로 바꿉니다.
	route := GraphRouteChat
	expression := ""
	if candidate, ok := ExtractCalculationExpression(question); ok {
		route = GraphRouteCalculator
		expression = candidate
	}

	return &assistantGraphState{
		Question:   question,
		History:    cloneMessages(input.History),
		Route:      route,
		Expression: expression,
	}, nil
}

func prepareGraphPromptVariables(_ context.Context, state *assistantGraphState) (map[string]any, error) {
	if state == nil {
		return nil, errors.New("assistant graph: state is required")
	}

	// Chapter 5 Chain과 같은 template 입력 형태를 재사용합니다.
	return chatChainInput(state.Question, state.History), nil
}

func runCalculatorGraphNode(ctx context.Context, state *assistantGraphState, trace *AssistantGraphResult) (*AssistantGraphResult, error) {
	if state == nil {
		return nil, errors.New("assistant graph: state is required")
	}

	calculation, err := tools.Calculate(ctx, tools.CalculatorInput{Expression: state.Expression})
	if err != nil {
		return nil, err
	}

	// calculator branch는 graph 결과를 즉시 반환하므로 downstream chat node가 실행되지 않습니다.
	result := &AssistantGraphResult{
		Route:       GraphRouteCalculator,
		Answer:      formatCalculationAnswer(calculation),
		Calculation: &calculation,
	}
	if trace != nil {
		*trace = *result
		return trace, nil
	}

	return result, nil
}

func ExtractCalculationExpression(question string) (string, bool) {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return "", false
	}

	// prefix가 있으면 routing 의도가 명확합니다. 편의를 위해 순수 산술식도 허용합니다.
	lower := strings.ToLower(trimmed)
	for _, prefix := range []string{"calculate:", "calc:", "calculate ", "calc "} {
		if strings.HasPrefix(lower, prefix) {
			expression := strings.TrimSpace(trimmed[len(prefix):])
			return expression, expression != ""
		}
	}

	if _, err := tools.Calculate(context.Background(), tools.CalculatorInput{Expression: trimmed}); err == nil {
		return trimmed, true
	}

	return "", false
}

func formatCalculationAnswer(calculation tools.CalculatorOutput) string {
	return fmt.Sprintf("%s = %s", calculation.Expression, formatFloat(calculation.Result))
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

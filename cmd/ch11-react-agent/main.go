package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	reactagent "github.com/sangjinsu/eino-learning/internal/llm/agent"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

const defaultQuestion = `calculator tool을 사용해 "12 * (3 + 4)"를 계산하고, ReAct Agent가 어떤 흐름으로 답했는지 짧게 설명해 주세요.`

var errToolCallingModelRequired = errors.New("OpenAI chat model does not support tool calling")

func main() {
	question := defaultQuestion
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed ReAct agent.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := run(ctx, cfg, question)
	if err != nil {
		log.Fatal(err)
	}

	printResult(os.Stdout, result)
}

func run(ctx context.Context, cfg llmopenai.Config, question string) (*reactagent.Result, error) {
	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}

	toolCallingModel, ok := chatModel.(model.ToolCallingChatModel)
	if !ok {
		return nil, errToolCallingModelRequired
	}

	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		return nil, err
	}

	service := reactagent.NewService(toolCallingModel, []einotool.BaseTool{calculatorTool})
	return service.Ask(ctx, question)
}

func printResult(w io.Writer, result *reactagent.Result) {
	fmt.Fprintln(w, "react agent:")
	fmt.Fprintln(w, "question -> ChatModel -> tool call -> ToolsNode -> ChatModel -> final answer")

	fmt.Fprintln(w)
	fmt.Fprintln(w, "question:")
	fmt.Fprintln(w, result.Question)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "available tools:")
	if len(result.AvailableTools) == 0 {
		fmt.Fprintln(w, "- none")
	}
	for _, name := range result.AvailableTools {
		fmt.Fprintf(w, "- %s\n", name)
	}
	fmt.Fprintf(w, "max step: %d\n", result.MaxStep)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "final answer:")
	fmt.Fprintln(w, result.Answer())
}

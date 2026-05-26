package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/llm/toolcalling"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func main() {
	expression := "12 * (7 + 3)"
	if len(os.Args) > 1 {
		expression = strings.Join(os.Args[1:], " ")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed tool calling.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	calculatorTool, err := tools.NewCalculatorTool()
	if err != nil {
		log.Fatal(err)
	}

	question := fmt.Sprintf(
		"Use the %s tool to calculate %q. Then answer with the expression and result.",
		tools.CalculatorToolName,
		expression,
	)
	chatService := toolcalling.NewService(chatModel)
	result, err := chatService.Ask(ctx, question, []einotool.BaseTool{calculatorTool})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("question:")
	fmt.Println(question)

	fmt.Println()
	fmt.Println("model tool calls:")
	if len(result.FirstResponse.ToolCalls) == 0 {
		fmt.Println("- none")
	}
	for _, call := range result.FirstResponse.ToolCalls {
		fmt.Printf("- id=%s name=%s args=%s\n", call.ID, call.Function.Name, call.Function.Arguments)
	}

	fmt.Println()
	fmt.Println("tool messages:")
	if len(result.ToolMessages) == 0 {
		fmt.Println("- none")
	}
	for _, msg := range result.ToolMessages {
		fmt.Printf("- role=%s tool_call_id=%s tool_name=%s content=%s\n", msg.Role, msg.ToolCallID, msg.ToolName, msg.Content)
	}

	fmt.Println()
	fmt.Println("final answer:")
	fmt.Println(result.Answer())
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/graph"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
)

func main() {
	questions := []string{
		"calculate: 12 * (7 + 3)",
		"How is Eino Graph different from Chain?",
	}
	if len(os.Args) > 1 {
		questions = []string{strings.Join(os.Args[1:], " ")}
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed Graph.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	service, err := graph.NewService(ctx, chatModel)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("graph:")
	fmt.Println("START -> route")
	fmt.Println("route -> calculator -> END")
	fmt.Println("route -> prepare_prompt -> ChatTemplate -> ChatModel -> END")

	history := []*schema.Message{
		schema.UserMessage("What did Chapter 5 cover?"),
		schema.AssistantMessage("It covered Chain as a linear pipeline.", nil),
	}

	for i, question := range questions {
		if i > 0 {
			fmt.Println()
			fmt.Println("---")
		}
		result, err := service.Run(ctx, graph.Input{
			Question: question,
			History:  history,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println()
		fmt.Printf("question: %s\n", question)
		fmt.Printf("selected route: %s\n", result.Route)

		switch result.Route {
		case graph.RouteCalculator:
			fmt.Println("calculator output:")
			fmt.Printf("- expression=%s\n", result.Calculation.Expression)
			fmt.Printf("- result=%g\n", result.Calculation.Result)
		case graph.RouteChat:
			fmt.Println("ChatTemplate output messages:")
			for j, msg := range result.PromptMessages {
				fmt.Printf("- message[%d] role=%s content=%s\n", j, msg.Role, msg.Content)
			}
			fmt.Println("ChatModel output:")
			fmt.Printf("- role=%s content=%s\n", result.ModelResponse.Role, result.ModelResponse.Content)
		}

		fmt.Println("final answer:")
		fmt.Println(result.Answer)
	}
}

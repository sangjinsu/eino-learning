package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm"
)

func main() {
	question := "How does Eino Chain compose components?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	cfg := llm.LoadOpenAIConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed Chain.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := llm.NewOpenAIChatModel(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	history := []*schema.Message{
		schema.UserMessage("What did Chapter 4 add?"),
		schema.AssistantMessage("It added model-backed tool calling with a calculator tool.", nil),
	}

	trace, err := llm.RunChatChainWithTrace(ctx, chatModel, llm.DefaultChatTemplate(), question, history)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("chain:")
	fmt.Println("input variables -> ChatTemplate -> prompt messages -> ChatModel -> assistant message")

	fmt.Println()
	fmt.Println("1. input variables:")
	if question, ok := trace.InputVariables["question"].(string); ok {
		fmt.Printf("- question=%s\n", question)
	}
	if tracedHistory, ok := trace.InputVariables["history"].([]*schema.Message); ok {
		for i, msg := range tracedHistory {
			fmt.Printf("- history[%d] role=%s content=%s\n", i, msg.Role, msg.Content)
		}
	}

	fmt.Println()
	fmt.Println("2. ChatTemplate output messages:")
	for i, msg := range trace.PromptMessages {
		fmt.Printf("- message[%d] role=%s content=%s\n", i, msg.Role, msg.Content)
	}

	fmt.Println()
	fmt.Println("3. ChatModel output:")
	fmt.Printf("- role=%s content=%s\n", trace.ModelResponse.Role, trace.ModelResponse.Content)

	fmt.Println()
	fmt.Println("final answer:")
	fmt.Println(trace.Answer())
}

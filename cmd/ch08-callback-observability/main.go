package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/observability"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

func main() {
	question := "Eino callback은 observability에 어떻게 도움이 되나요?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed callback observability.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	history := []*schema.Message{
		schema.UserMessage("Chapter 7에서는 무엇을 다뤘나요?"),
		schema.AssistantMessage("StreamReader를 사용한 streaming 흐름을 다뤘습니다.", nil),
	}

	result, err := observability.RunObservableChatChain(ctx, chatModel, prompting.DefaultChatTemplate(), question, history)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("callback observability:")
	fmt.Println("question + history -> ChatTemplate -> ChatModel")
	fmt.Println()
	fmt.Println("callback events:")
	for i, event := range result.Events {
		fmt.Printf("- event[%d] timing=%s name=%s component=%s summary=%s\n",
			i, event.Timing, event.Name, event.Component, event.Summary)
	}

	fmt.Println()
	fmt.Println("final answer:")
	fmt.Println(result.Answer)
}

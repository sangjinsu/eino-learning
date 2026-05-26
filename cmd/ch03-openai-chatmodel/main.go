package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sangjinsu/eino-learning/internal/llm/chat"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
)

func main() {
	question := "What does Eino ChatModel do?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	if !llmopenai.IntegrationEnabled() {
		fmt.Println("OpenAI integration is disabled.")
		fmt.Println("Set RUN_EINO_INTEGRATION=1 and OPENAI_API_KEY to run this example.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, llmopenai.LoadConfigFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	service := chat.NewService(chatModel)
	answer, err := service.Ask(ctx, question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}

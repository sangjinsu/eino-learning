package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm/chat"
)

func main() {
	question := "What is Eino?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	chatModel := fake.NewChatModel("Eino helps build testable LLM applications in Go.")
	service := chat.NewService(chatModel)

	answer, err := service.Ask(context.Background(), question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}

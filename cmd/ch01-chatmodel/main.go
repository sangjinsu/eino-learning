package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm"
)

func main() {
	question := "What is Eino?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	chatModel := fake.NewChatModel("Eino helps build testable LLM applications in Go.")
	service := llm.NewChatService(chatModel)

	answer, err := service.Ask(context.Background(), question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}

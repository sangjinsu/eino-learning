package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm"
)

func main() {
	question := "How does ChatTemplate work?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	chatModel := fake.NewChatModel("ChatTemplate turns variables into ordered chat messages for a ChatModel.")
	service := llm.NewChatService(chatModel)
	history := []*schema.Message{
		schema.UserMessage("What did chapter 1 cover?"),
		schema.AssistantMessage("It covered fake ChatModel basics.", nil),
	}

	answer, err := service.AskWithHistory(context.Background(), question, history)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("messages sent to model:")
	for _, msg := range chatModel.LastInput() {
		fmt.Printf("- %s: %s\n", msg.Role, msg.Content)
	}
	fmt.Println()
	fmt.Println("answer:")
	fmt.Println(answer)
}

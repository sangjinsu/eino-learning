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

	"github.com/cloudwego/eino/schema"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/llm/streaming"
)

func main() {
	question := "Explain Eino streaming in one short paragraph."
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed Streaming.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	service := streaming.NewService(chatModel)

	history := []*schema.Message{
		schema.UserMessage("What did Chapter 6 cover?"),
		schema.AssistantMessage("It covered Graph branching with calculator and chat paths.", nil),
	}

	reader, err := service.StreamWithHistory(ctx, question, history)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Println("streaming:")
	fmt.Println("question -> ChatTemplate -> ChatModel.Stream -> StreamReader.Recv loop")
	fmt.Println()
	fmt.Println("stream chunks:")

	var answer strings.Builder
	chunkCount := 0
	for {
		chunk, err := reader.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if chunk == nil || chunk.Content == "" {
			continue
		}

		fmt.Print(chunk.Content)
		answer.WriteString(chunk.Content)
		chunkCount++
	}

	fmt.Println()
	fmt.Println()
	fmt.Printf("received chunks: %d\n", chunkCount)
	fmt.Println("final answer:")
	fmt.Println(answer.String())
}

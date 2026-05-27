package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sangjinsu/eino-learning/internal/mcpdemo"
)

func main() {
	// stdout은 MCP stdio transport가 사용하므로 일반 안내 문구를 출력하지 않습니다.
	server := mcpdemo.NewServer()
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

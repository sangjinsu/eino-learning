package mcpdemo

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

const (
	ServerName         = "eino-learning-mcp"
	ServerVersion      = "v0.1.0"
	ChapterResourceURI = "eino-learning://chapters/mcp"
)

const chapterResourceText = `MCP basics for eino-learning:
- tools expose executable actions through a standard protocol.
- resources expose read-only context that clients can discover and read.
- prompts expose reusable user-triggered prompt templates.
This chapter starts with a local stdio MCP server, a safe calculator tool, and a read-only learning resource.`

// calculatorInput은 MCP SDK의 JSON schema 추론에 맞춘 얇은 입력 타입입니다.
// 기존 tools.CalculatorInput의 jsonschema tag는 Eino tool용 형식이라 MCP SDK와 분리합니다.
type calculatorInput struct {
	Expression string `json:"expression"`
}

// NewServer는 Chapter 10에서 사용할 학습용 MCP server를 구성합니다.
func NewServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: ServerName, Version: ServerVersion}, nil)

	// Chapter 04의 safe calculator 로직을 MCP tool로 다시 노출합니다.
	mcp.AddTool(server, &mcp.Tool{
		Name:        tools.CalculatorToolName,
		Title:       "Calculator",
		Description: "Evaluate a safe arithmetic expression using +, -, *, /, and parentheses.",
	}, calculateTool)
	// Resource는 model 입력이 아니라 client가 읽을 수 있는 context입니다.
	server.AddResource(&mcp.Resource{
		URI:         ChapterResourceURI,
		Name:        "mcp-basics",
		Title:       "MCP Basics",
		Description: "A short explanation of MCP tools, resources, and prompts for Chapter 10.",
		MIMEType:    "text/plain",
	}, readChapterResource)
	return server
}

func calculateTool(ctx context.Context, _ *mcp.CallToolRequest, input calculatorInput) (*mcp.CallToolResult, tools.CalculatorOutput, error) {
	output, err := tools.Calculate(ctx, tools.CalculatorInput{Expression: input.Expression})
	return nil, output, err
}

// readChapterResource는 허용된 URI 하나만 읽기 전용 resource로 반환합니다.
func readChapterResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := ""
	if req != nil && req.Params != nil {
		uri = req.Params.URI
	}
	if uri != ChapterResourceURI {
		return nil, mcp.ResourceNotFoundError(uri)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      ChapterResourceURI,
				MIMEType: "text/plain",
				Text:     chapterResourceText,
			},
		},
	}, nil
}

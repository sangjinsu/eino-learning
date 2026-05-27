package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sangjinsu/eino-learning/internal/mcpdemo"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

const demoExpression = "2 + 3 * 4"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := run(ctx, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, out io.Writer) error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	serverCmd := exec.CommandContext(ctx, "go", "run", "./cmd/ch10-mcp-server")
	serverCmd.Dir = root
	serverCmd.Stderr = os.Stderr

	return runWithServerCommand(ctx, out, serverCmd, "go run ./cmd/ch10-mcp-server")
}

func runWithServerCommand(ctx context.Context, out io.Writer, serverCmd *exec.Cmd, serverCommandText string) error {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "eino-learning-mcp-demo-client",
		Version: "v0.1.0",
	}, nil)

	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: serverCmd}, nil)
	if err != nil {
		return fmt.Errorf("connect to MCP server: %w", err)
	}
	defer session.Close()

	toolResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tools.CalculatorToolName,
		Arguments: map[string]any{"expression": demoExpression},
	})
	if err != nil {
		return fmt.Errorf("call calculator tool: %w", err)
	}
	if toolResult.IsError {
		return fmt.Errorf("calculator tool returned error: %v", toolResult.GetError())
	}

	calculatorOutput, err := decodeCalculatorOutput(toolResult.StructuredContent)
	if err != nil {
		return err
	}

	resourceResult, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: mcpdemo.ChapterResourceURI})
	if err != nil {
		return fmt.Errorf("read MCP chapter resource: %w", err)
	}

	resourceText, err := firstResourceText(resourceResult)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "mcp demo:")
	fmt.Fprintf(out, "server command: %s\n", serverCommandText)
	fmt.Fprintf(out, "tool call: %s expression=%q\n", tools.CalculatorToolName, demoExpression)
	fmt.Fprintf(out, "tool result: expression=%q result=%g\n", calculatorOutput.Expression, calculatorOutput.Result)
	fmt.Fprintf(out, "resource read: %s\n", mcpdemo.ChapterResourceURI)
	fmt.Fprintln(out, "resource text:")
	fmt.Fprintln(out, resourceText)
	return nil
}

func decodeCalculatorOutput(content any) (tools.CalculatorOutput, error) {
	if content == nil {
		return tools.CalculatorOutput{}, errors.New("calculator tool did not return structured content")
	}

	data, err := json.Marshal(content)
	if err != nil {
		return tools.CalculatorOutput{}, fmt.Errorf("marshal calculator output: %w", err)
	}

	var output tools.CalculatorOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return tools.CalculatorOutput{}, fmt.Errorf("unmarshal calculator output: %w", err)
	}
	return output, nil
}

func firstResourceText(result *mcp.ReadResourceResult) (string, error) {
	if result == nil || len(result.Contents) == 0 {
		return "", errors.New("MCP chapter resource returned no contents")
	}
	if result.Contents[0].Text == "" {
		return "", errors.New("MCP chapter resource returned empty text")
	}
	return result.Contents[0].Text, nil
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	start := dir

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("check go.mod in %s: %w", dir, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		dir = parent
	}
}

package mcpdemo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestServerExposesCalculatorTool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, cleanup := connectTestSession(t, ctx, NewServer())
	defer cleanup()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tools.CalculatorToolName,
		Arguments: map[string]any{"expression": "2 + 3 * 4"},
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool returned tool error: %#v", result.Content)
	}

	var output tools.CalculatorOutput
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured output: %v", err)
	}
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("unmarshal structured output: %v", err)
	}
	if output.Expression != "2 + 3 * 4" {
		t.Fatalf("expression = %q, want original expression", output.Expression)
	}
	if output.Result != 14 {
		t.Fatalf("result = %v, want 14", output.Result)
	}
}

func TestServerExposesMCPChapterResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, cleanup := connectTestSession(t, ctx, NewServer())
	defer cleanup()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: ChapterResourceURI})
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("contents length = %d, want 1", len(result.Contents))
	}

	content := result.Contents[0]
	if content.URI != ChapterResourceURI {
		t.Fatalf("resource URI = %q, want %q", content.URI, ChapterResourceURI)
	}
	if content.MIMEType != "text/plain" {
		t.Fatalf("MIMEType = %q, want text/plain", content.MIMEType)
	}
	if !strings.Contains(content.Text, "tools") || !strings.Contains(content.Text, "resources") {
		t.Fatalf("resource text = %q, want MCP tools/resources summary", content.Text)
	}
}

func connectTestSession(t *testing.T, ctx context.Context, server *mcp.Server) (*mcp.ClientSession, func()) {
	t.Helper()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "mcpdemo-test", Version: "v0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	cleanup := func() {
		_ = session.Close()
		select {
		case err := <-errCh:
			if err != nil && !strings.Contains(err.Error(), "closed") && !strings.Contains(err.Error(), "canceled") {
				t.Fatalf("server Run returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("server Run did not stop after session close")
		}
	}

	return session, cleanup
}

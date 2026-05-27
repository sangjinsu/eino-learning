package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRunDemoStartsMCPServerAndPrintsResults(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	serverCmd := exec.CommandContext(ctx, "go", "run", "./cmd/ch10-mcp-server")
	serverCmd.Dir = root

	var out bytes.Buffer
	if err := runWithServerCommand(ctx, &out, serverCmd, "go run ./cmd/ch10-mcp-server"); err != nil {
		t.Fatalf("run demo: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"mcp demo:",
		"server command: go run ./cmd/ch10-mcp-server",
		"tool call: calculator expression=\"2 + 3 * 4\"",
		"tool result: expression=\"2 + 3 * 4\" result=14",
		"resource read: eino-learning://chapters/mcp",
		"MCP basics for eino-learning",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\noutput:\n%s", want, got)
		}
	}
}

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	reactagent "github.com/sangjinsu/eino-learning/internal/llm/agent"
	"github.com/sangjinsu/eino-learning/internal/tools"
)

func TestPrintResultShowsQuestionToolsAndFinalAnswer(t *testing.T) {
	result := &reactagent.Result{
		Question:       "Use calculator to solve 2 + 3 * 4.",
		FinalResponse:  schema.AssistantMessage("2 + 3 * 4 = 14.", nil),
		AvailableTools: []string{tools.CalculatorToolName},
		MaxStep:        12,
	}

	var out bytes.Buffer
	printResult(&out, result)

	got := out.String()
	for _, want := range []string{
		"react agent:",
		"question:",
		"Use calculator to solve 2 + 3 * 4.",
		"available tools:",
		"- calculator",
		"max step: 12",
		"final answer:",
		"2 + 3 * 4 = 14.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\noutput:\n%s", want, got)
		}
	}
}

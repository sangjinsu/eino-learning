package graph

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

type messageWant struct {
	role    schema.RoleType
	content string
}

func assertMessages(t *testing.T, got []*schema.Message, want []messageWant) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("message length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Role != want[i].role {
			t.Fatalf("message[%d].Role = %q, want %q", i, got[i].Role, want[i].role)
		}
		if got[i].Content != want[i].content {
			t.Fatalf("message[%d].Content = %q, want %q", i, got[i].Content, want[i].content)
		}
	}
}

package openai

import (
	"context"
	"errors"
	"testing"
)

func TestNewChatModelValidatesConfigBeforeCreatingProvider(t *testing.T) {
	_, err := NewChatModel(context.Background(), Config{
		Model: DefaultModel,
	})
	if !errors.Is(err, ErrAPIKeyRequired) {
		t.Fatalf("NewChatModel() error = %v, want %v", err, ErrAPIKeyRequired)
	}
}

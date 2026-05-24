package llm

import (
	"context"
	"errors"
	"testing"
)

func TestNewOpenAIChatModelValidatesConfigBeforeCreatingProvider(t *testing.T) {
	_, err := NewOpenAIChatModel(context.Background(), OpenAIConfig{
		Model: DefaultOpenAIModel,
	})
	if !errors.Is(err, ErrOpenAIAPIKeyRequired) {
		t.Fatalf("NewOpenAIChatModel() error = %v, want %v", err, ErrOpenAIAPIKeyRequired)
	}
}

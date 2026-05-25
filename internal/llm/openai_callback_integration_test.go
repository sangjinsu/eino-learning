package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

func TestOpenAIObservableChatChainIntegration(t *testing.T) {
	if !OpenAIIntegrationEnabled() {
		t.Skip("set RUN_EINO_INTEGRATION=1 to run OpenAI callback integration test")
	}

	cfg := LoadOpenAIConfigFromEnv()
	if strings.TrimSpace(cfg.APIKey) == "" {
		t.Skip("set OPENAI_API_KEY to run OpenAI callback integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	chatModel, err := NewOpenAIChatModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOpenAIChatModel returned error: %v", err)
	}

	result, err := RunObservableChatChain(ctx, chatModel, DefaultChatTemplate(), "한 문장으로 Eino callback은 무엇을 관찰하나요?", []*schema.Message{
		schema.UserMessage("Chapter 7에서는 무엇을 다뤘나요?"),
		schema.AssistantMessage("Streaming을 다뤘습니다.", nil),
	})
	if err != nil {
		t.Fatalf("RunObservableChatChain returned error: %v", err)
	}
	if strings.TrimSpace(result.Answer) == "" {
		t.Fatal("answer is blank")
	}
	if len(result.Events) == 0 {
		t.Fatal("callback events are empty")
	}

	_ = findCallbackEvent(t, result.Events, CallbackTimingStart, "ChatTemplate")
	_ = findCallbackEvent(t, result.Events, CallbackTimingEnd, "ChatModel")
}

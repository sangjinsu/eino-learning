package rag

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/fake"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

func TestServiceRetrievesContextAndReturnsSources(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		{
			ID:      "streaming",
			Content: "Chapter 07 Streaming explains StreamReader.Recv loops and io.EOF handling.",
			MetaData: map[string]any{
				MetaKeyTitle:  "Chapter 07 Streaming",
				MetaKeySource: "testdata/docs/ch09-rag/streaming.md",
			},
		},
		{
			ID:      "callbacks",
			Content: "Chapter 08 Callback records start, end, and error lifecycle events for Eino components.",
			MetaData: map[string]any{
				MetaKeyTitle:  "Chapter 08 Callback",
				MetaKeySource: "testdata/docs/ch09-rag/callback.md",
			},
		},
	}
	chatModel := fake.NewChatModel("Callback events can be used as RAG context.")
	service, err := NewService(ctx, NewInMemoryRetriever(docs), chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	got, err := service.Ask(ctx, "callback lifecycle event는 무엇을 기록하나요?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if got.Answer != "Callback events can be used as RAG context." {
		t.Fatalf("answer = %q, want fake model response", got.Answer)
	}
	if len(got.Sources) != 1 {
		t.Fatalf("sources length = %d, want 1: %#v", len(got.Sources), got.Sources)
	}
	if got.Sources[0].ID != "callbacks" {
		t.Fatalf("source ID = %q, want callbacks", got.Sources[0].ID)
	}
	if got.Sources[0].Title != "Chapter 08 Callback" {
		t.Fatalf("source title = %q, want metadata title", got.Sources[0].Title)
	}
	if got.Sources[0].Source != "testdata/docs/ch09-rag/callback.md" {
		t.Fatalf("source path = %q, want metadata source", got.Sources[0].Source)
	}
	if len(got.RetrievedDocuments) != 1 || got.RetrievedDocuments[0].ID != "callbacks" {
		t.Fatalf("retrieved documents = %#v, want callback document", got.RetrievedDocuments)
	}

	prompt := lastUserMessageContent(t, got.PromptMessages)
	assertContains(t, prompt, "Chapter 08 Callback records start, end, and error lifecycle events")
	assertContains(t, prompt, "callback lifecycle event는 무엇을 기록하나요?")
	assertContains(t, prompt, "testdata/docs/ch09-rag/callback.md")

	modelPrompt := lastUserMessageContent(t, chatModel.LastInput())
	if modelPrompt != prompt {
		t.Fatalf("model prompt = %q, want captured prompt %q", modelPrompt, prompt)
	}
}

func TestInMemoryRetrieverHonorsTopKAndScoresByKeywordOverlap(t *testing.T) {
	ctx := context.Background()
	retr := NewInMemoryRetriever([]*schema.Document{
		{ID: "chain", Content: "Chain connects ChatTemplate and ChatModel in a linear pipeline."},
		{ID: "graph", Content: "Graph uses route branches for calculator and chat model paths."},
		{ID: "callback", Content: "Callback observes component lifecycle events."},
	})

	got, err := retr.Retrieve(ctx, "graph route branch calculator", retriever.WithTopK(1))
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("retrieved length = %d, want 1: %#v", len(got), got)
	}
	if got[0].ID != "graph" {
		t.Fatalf("first document ID = %q, want graph", got[0].ID)
	}
	if got[0].Score() <= 0 {
		t.Fatalf("score = %v, want positive keyword score", got[0].Score())
	}
}

func TestInMemoryRetrieverUsesChapter09TestdataDocuments(t *testing.T) {
	ctx := context.Background()
	retr := NewInMemoryRetriever([]*schema.Document{
		testdataDocument(t, "chapter08", "Chapter 08 Callback Observability", "chapter08-callback-observability.md"),
		testdataDocument(t, "chapter09", "Chapter 09 RAG Basics", "chapter09-rag-basics.txt"),
		testdataDocument(t, "chapter06", "Chapter 06 Graph", "chapter06-graph.md"),
	})

	got, err := retr.Retrieve(ctx, "callback observability events rag context", retriever.WithTopK(1))
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("retrieved length = %d, want 1: %#v", len(got), got)
	}
	if got[0].ID != "chapter08" {
		t.Fatalf("first document ID = %q, want chapter08", got[0].ID)
	}
	if got[0].MetaData[MetaKeySource] != "chapter08-callback-observability.md" {
		t.Fatalf("source metadata = %v, want testdata source", got[0].MetaData[MetaKeySource])
	}
}

func TestInMemoryRetrieverFindsCallbackDocumentForKoreanRAGQuestion(t *testing.T) {
	ctx := context.Background()
	retr := NewInMemoryRetriever([]*schema.Document{
		testdataDocument(t, "chapter08", "Chapter 08 Callback Observability", "chapter08-callback-observability.md"),
		testdataDocument(t, "chapter09", "Chapter 09 RAG Basics", "chapter09-rag-basics.txt"),
		testdataDocument(t, "chapter06", "Chapter 06 Graph", "chapter06-graph.md"),
	})

	got, err := retr.Retrieve(ctx, "Chapter 8 callback은 RAG에서 어떤 흐름을 관찰하나요?", retriever.WithTopK(2))
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("retrieved no documents")
	}
	if got[0].ID != "chapter08" {
		t.Fatalf("first document ID = %q, want chapter08: %#v", got[0].ID, got)
	}
}

func TestInMemoryRetrieverFindsStreamingDocumentForKoreanRAGQuestion(t *testing.T) {
	ctx := context.Background()
	retr := NewInMemoryRetriever([]*schema.Document{
		testdataDocument(t, "chapter07", "Chapter 07 Streaming", "chapter07-streaming.md"),
		testdataDocument(t, "chapter08", "Chapter 08 Callback Observability", "chapter08-callback-observability.md"),
		testdataDocument(t, "chapter09", "Chapter 09 RAG Basics", "chapter09-rag-basics.txt"),
		testdataDocument(t, "chapter06", "Chapter 06 Graph", "chapter06-graph.md"),
	})

	got, err := retr.Retrieve(ctx, "Chapter 7 streaming은 RAG와 어떻게 연결될 수 있나요?", retriever.WithTopK(2))
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("retrieved no documents")
	}
	if got[0].ID != "chapter07" {
		t.Fatalf("first document ID = %q, want chapter07: %#v", got[0].ID, got)
	}
}

func TestServiceRejectsBlankQuestionBeforeRetrievalOrModel(t *testing.T) {
	ctx := context.Background()
	spy := &spyRetriever{}
	chatModel := fake.NewChatModel("unused")
	service, err := NewService(ctx, spy, chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	_, err = service.Ask(ctx, " \t\n ")
	if !errors.Is(err, prompting.ErrBlankQuestion) {
		t.Fatalf("Ask error = %v, want %v", err, prompting.ErrBlankQuestion)
	}
	if spy.calls != 0 {
		t.Fatalf("retriever calls = %d, want 0", spy.calls)
	}
	if len(chatModel.LastInput()) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(chatModel.LastInput()))
	}
}

func TestServiceReturnsErrorWhenNoRelevantDocuments(t *testing.T) {
	ctx := context.Background()
	chatModel := fake.NewChatModel("unused")
	service, err := NewService(ctx, NewInMemoryRetriever(nil), chatModel)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	_, err = service.Ask(ctx, "RAG가 무엇인가요?")
	if !errors.Is(err, ErrNoRelevantDocuments) {
		t.Fatalf("Ask error = %v, want %v", err, ErrNoRelevantDocuments)
	}
	if len(chatModel.LastInput()) != 0 {
		t.Fatalf("model was called with %d messages, want 0", len(chatModel.LastInput()))
	}
}

type spyRetriever struct {
	calls int
}

func (r *spyRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	_ = ctx
	_ = query
	_ = opts

	r.calls++
	return []*schema.Document{{ID: "unused", Content: "unused"}}, nil
}

func lastUserMessageContent(t *testing.T, messages []*schema.Message) string {
	t.Helper()

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User {
			return messages[i].Content
		}
	}

	t.Fatalf("user message not found in %#v", messages)
	return ""
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Fatalf("%q does not contain %q", got, want)
	}
}

func testdataDocument(t *testing.T, id string, title string, source string) *schema.Document {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "docs", "ch09-rag", source))
	if err != nil {
		t.Fatalf("read testdata document %s: %v", source, err)
	}

	return &schema.Document{
		ID:      id,
		Content: string(content),
		MetaData: map[string]any{
			MetaKeyTitle:  title,
			MetaKeySource: source,
		},
	}
}

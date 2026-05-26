package main

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/rag"
)

func TestDocumentContextPreviewTrimsWhitespaceAndLimitsLength(t *testing.T) {
	doc := &schema.Document{
		ID:      "callbacks",
		Content: "  Eino callback records component lifecycle events so RAG answers can be traced back to retrieval and model execution.  ",
		MetaData: map[string]any{
			"title":  "Chapter 08 Callback",
			"source": "testdata/docs/ch09-rag/callbacks.md",
		},
	}

	got := documentContextPreview(doc, 64)

	if !strings.Contains(got, "Chapter 08 Callback") {
		t.Fatalf("preview = %q, want title", got)
	}
	if !strings.Contains(got, "testdata/docs/ch09-rag/callbacks.md") {
		t.Fatalf("preview = %q, want source", got)
	}
	if strings.Contains(got, "  ") {
		t.Fatalf("preview = %q, want collapsed whitespace", got)
	}
	if !strings.Contains(got, "...") {
		t.Fatalf("preview = %q, want truncation marker", got)
	}
}

func TestMetadataStringFallsBackToDefault(t *testing.T) {
	doc := &schema.Document{
		ID:       "rag",
		MetaData: map[string]any{"title": "RAG Basics"},
	}

	if got := metadataString(doc, "title", "untitled"); got != "RAG Basics" {
		t.Fatalf("metadataString title = %q, want RAG Basics", got)
	}
	if got := metadataString(doc, "source", "unknown"); got != "unknown" {
		t.Fatalf("metadataString source = %q, want unknown", got)
	}
}

func TestDefaultQuestionRetrievesSampleDocument(t *testing.T) {
	docs, err := loadDocuments(defaultDocsDir)
	if err != nil {
		t.Fatalf("loadDocuments returned error: %v", err)
	}

	retriever := rag.NewInMemoryRetriever(docs)
	got, err := retriever.Retrieve(context.Background(), defaultQuestion)
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("default question retrieved no documents")
	}
	if got[0].ID != "chapter09-rag-basics" {
		t.Fatalf("first document ID = %q, want chapter09-rag-basics", got[0].ID)
	}
}

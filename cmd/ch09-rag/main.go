package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/eino/schema"
	llmopenai "github.com/sangjinsu/eino-learning/internal/llm/openai"
	"github.com/sangjinsu/eino-learning/internal/llm/rag"
)

const (
	defaultDocsDir  = "testdata/docs/ch09-rag"
	defaultQuestion = "RAG는 Eino 학습 예제에서 어떤 문제를 해결하나요?"
)

func main() {
	question := defaultQuestion
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	cfg := llmopenai.LoadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Println("OpenAI API key is not configured.")
		fmt.Println("Set OPENAI_API_KEY in your shell or .env to run model-backed RAG.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	docs, err := loadDocuments(defaultDocsDir)
	if err != nil {
		log.Fatal(err)
	}
	ragRetriever := rag.NewInMemoryRetriever(docs)

	chatModel, err := llmopenai.NewChatModel(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	service, err := rag.NewService(ctx, ragRetriever, chatModel)
	if err != nil {
		log.Fatal(err)
	}

	result, err := service.Ask(ctx, question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("rag:")
	fmt.Println("question -> Retriever -> context prompt -> ChatModel -> answer + sources")
	fmt.Println()
	fmt.Println("retrieved sources:")
	printSources(result.Sources)
	fmt.Println()
	fmt.Println("prompt context summary:")
	printPromptContextSummary(result.PromptMessages, result.RetrievedDocuments)
	fmt.Println()
	fmt.Println("final answer:")
	fmt.Println(result.Answer)
}

func printSources(sources []rag.Source) {
	if len(sources) == 0 {
		fmt.Println("- none")
		return
	}

	for i, source := range sources {
		title := fallbackString(source.Title, "untitled")
		path := fallbackString(source.Source, "unknown")
		if source.Score > 0 {
			fmt.Printf("- source[%d] title=%s source=%s score=%.2f\n", i, title, path, source.Score)
			continue
		}
		fmt.Printf("- source[%d] title=%s source=%s\n", i, title, path)
	}
}

func loadDocuments(dir string) ([]*schema.Document, error) {
	resolvedDir, err := resolveDocsDir(dir)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(resolvedDir)
	if err != nil {
		return nil, fmt.Errorf("read RAG docs dir: %w", err)
	}

	docs := make([]*schema.Document, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isTextDocument(entry.Name()) {
			continue
		}

		path := filepath.Join(resolvedDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read RAG doc %s: %w", path, err)
		}

		id := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		text := strings.TrimSpace(string(content))
		source := filepath.Join(dir, entry.Name())
		docs = append(docs, &schema.Document{
			ID:      id,
			Content: text,
			MetaData: map[string]any{
				rag.MetaKeyTitle:  documentTitle(text, id),
				rag.MetaKeySource: filepath.ToSlash(source),
			},
		})
	}

	return docs, nil
}

func resolveDocsDir(dir string) (string, error) {
	if isDir(dir) {
		return dir, nil
	}

	current := "."
	for range 5 {
		current = filepath.Join(current, "..")
		candidate := filepath.Join(current, dir)
		if isDir(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("read RAG docs dir: %w", os.ErrNotExist)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isTextDocument(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".txt":
		return true
	default:
		return false
	}
}

func documentTitle(content string, fallback string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
		if line != "" {
			return line
		}
	}
	return fallback
}

func printPromptContextSummary(messages []*schema.Message, docs []*schema.Document) {
	fmt.Printf("- prompt messages=%d retrieved documents=%d\n", len(messages), len(docs))
	for i, doc := range docs {
		fmt.Printf("- context[%d] %s\n", i, documentContextPreview(doc, 160))
	}
}

func documentContextPreview(doc *schema.Document, maxRunes int) string {
	if doc == nil {
		return "title=untitled source=unknown content="
	}

	title := metadataString(doc, "title", "untitled")
	source := metadataString(doc, "source", fallbackString(doc.ID, "unknown"))
	content := collapseWhitespace(doc.Content)
	if maxRunes > 0 && utf8.RuneCountInString(content) > maxRunes {
		runes := []rune(content)
		content = string(runes[:maxRunes]) + "..."
	}

	return fmt.Sprintf("title=%s source=%s content=%s", title, source, content)
}

func metadataString(doc *schema.Document, key string, fallback string) string {
	if doc == nil || doc.MetaData == nil {
		return fallback
	}

	value, ok := doc.MetaData[key]
	if !ok {
		return fallback
	}

	text, ok := value.(string)
	if !ok {
		return fallback
	}

	return fallbackString(text, fallback)
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

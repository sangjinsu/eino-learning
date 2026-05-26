package rag

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/sangjinsu/eino-learning/internal/llm/prompting"
)

const (
	MetaKeyTitle  = "title"
	MetaKeySource = "source"

	DefaultTopK = 2
)

var (
	ErrRetrieverRequired   = errors.New("rag service: retriever is required")
	ErrModelRequired       = errors.New("rag service: model is required")
	ErrTemplateRequired    = errors.New("rag service: template is required")
	ErrNoRelevantDocuments = errors.New("rag service: no relevant documents found")
)

type Source struct {
	ID     string
	Title  string
	Source string
	Score  float64
}

type Result struct {
	Answer             string
	Sources            []Source
	RetrievedDocuments []*schema.Document
	PromptMessages     []*schema.Message
	ModelResponse      *schema.Message
}

type Service struct {
	retriever retriever.Retriever
	model     model.BaseChatModel
	template  prompt.ChatTemplate
	topK      int
}

type InMemoryRetriever struct {
	docs        []*schema.Document
	defaultTopK int
}

func NewService(ctx context.Context, retriever retriever.Retriever, chatModel model.BaseChatModel) (*Service, error) {
	return NewServiceWithTemplate(ctx, retriever, chatModel, DefaultRAGTemplate())
}

func NewServiceWithTemplate(_ context.Context, retriever retriever.Retriever, chatModel model.BaseChatModel, template prompt.ChatTemplate) (*Service, error) {
	if retriever == nil {
		return nil, ErrRetrieverRequired
	}
	if chatModel == nil {
		return nil, ErrModelRequired
	}
	if template == nil {
		return nil, ErrTemplateRequired
	}

	return &Service{
		retriever: retriever,
		model:     chatModel,
		template:  template,
		topK:      DefaultTopK,
	}, nil
}

func DefaultRAGTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(prompting.DefaultSystemPrompt+" Answer only from the retrieved context. If the context is insufficient, say so clearly."),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("Use the retrieved context below to answer the question.\n\nRetrieved context:\n{context}\n\nQuestion: {question}"),
	)
}

func (s *Service) Ask(ctx context.Context, question string) (*Result, error) {
	return s.AskWithHistory(ctx, question, nil)
}

func (s *Service) AskWithHistory(ctx context.Context, question string, history []*schema.Message) (*Result, error) {
	if prompting.IsBlankQuestion(question) {
		return nil, prompting.ErrBlankQuestion
	}
	if s == nil || s.retriever == nil {
		return nil, ErrRetrieverRequired
	}
	if s.model == nil {
		return nil, ErrModelRequired
	}
	if s.template == nil {
		return nil, ErrTemplateRequired
	}

	docs, err := s.retriever.Retrieve(ctx, strings.TrimSpace(question), retriever.WithTopK(s.topK))
	if err != nil {
		return nil, fmt.Errorf("retrieve RAG context: %w", err)
	}
	if len(docs) == 0 {
		return nil, ErrNoRelevantDocuments
	}

	vars := map[string]any{
		"question": strings.TrimSpace(question),
		"context":  formatContext(docs),
	}
	if len(history) > 0 {
		vars["history"] = history
	}

	messages, err := s.template.Format(ctx, vars)
	if err != nil {
		return nil, fmt.Errorf("format RAG prompt: %w", err)
	}

	message, err := s.model.Generate(ctx, messages)
	if err != nil {
		return &Result{
			Sources:            sourcesFromDocuments(docs),
			RetrievedDocuments: cloneDocuments(docs),
			PromptMessages:     prompting.CloneMessages(messages),
		}, fmt.Errorf("generate RAG answer: %w", err)
	}
	if message == nil {
		return nil, errors.New("generate RAG answer: empty response")
	}

	return &Result{
		Answer:             message.Content,
		Sources:            sourcesFromDocuments(docs),
		RetrievedDocuments: cloneDocuments(docs),
		PromptMessages:     prompting.CloneMessages(messages),
		ModelResponse:      message,
	}, nil
}

func NewInMemoryRetriever(docs []*schema.Document) *InMemoryRetriever {
	return &InMemoryRetriever{
		docs:        cloneDocuments(docs),
		defaultTopK: DefaultTopK,
	}
}

func (r *InMemoryRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	_ = ctx

	if r == nil {
		return nil, ErrRetrieverRequired
	}

	options := retriever.GetCommonOptions(&retriever.Options{TopK: &r.defaultTopK}, opts...)
	topK := r.defaultTopK
	if options.TopK != nil {
		topK = *options.TopK
	}
	if topK <= 0 {
		return nil, nil
	}

	queryWeights := tokenWeights(query)
	if len(queryWeights) == 0 {
		return nil, nil
	}

	scored := make([]*schema.Document, 0, len(r.docs))
	for _, doc := range r.docs {
		score := scoreDocument(queryWeights, doc)
		if options.ScoreThreshold != nil && score < *options.ScoreThreshold {
			continue
		}
		if score <= 0 {
			continue
		}

		copied := cloneDocument(doc)
		copied.WithScore(score)
		scored = append(scored, copied)
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score() == scored[j].Score() {
			return scored[i].ID < scored[j].ID
		}
		return scored[i].Score() > scored[j].Score()
	})

	if len(scored) > topK {
		scored = scored[:topK]
	}

	return scored, nil
}

func scoreDocument(queryWeights map[string]int, doc *schema.Document) float64 {
	if doc == nil {
		return 0
	}

	docWeights := tokenWeights(doc.ID + " " + metadataString(doc.MetaData, MetaKeyTitle) + " " + doc.Content)
	score := 0
	for token, weight := range queryWeights {
		if docWeight := docWeights[token]; docWeight > 0 {
			score += weight * docWeight
		}
	}

	return float64(score)
}

func tokenWeights(text string) map[string]int {
	tokens := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	weights := make(map[string]int)
	for _, token := range tokens {
		token = normalizeToken(token)
		if token == "" {
			continue
		}
		weights[token]++
	}

	return weights
}

func normalizeToken(token string) string {
	token = strings.TrimSpace(token)
	if len([]rune(token)) <= 2 {
		return ""
	}
	if strings.HasSuffix(token, "ies") && len(token) > 4 {
		return strings.TrimSuffix(token, "ies") + "y"
	}
	if strings.HasSuffix(token, "es") && len(token) > 3 {
		return strings.TrimSuffix(token, "es")
	}
	if strings.HasSuffix(token, "s") && len(token) > 3 {
		return strings.TrimSuffix(token, "s")
	}
	return token
}

func formatContext(docs []*schema.Document) string {
	var b strings.Builder
	for i, doc := range docs {
		if i > 0 {
			b.WriteString("\n\n")
		}

		source := sourceFromDocument(doc)
		fmt.Fprintf(&b, "[%d] %s\nSource: %s\n%s", i+1, source.Title, source.Source, doc.Content)
	}

	return b.String()
}

func sourcesFromDocuments(docs []*schema.Document) []Source {
	sources := make([]Source, 0, len(docs))
	for _, doc := range docs {
		sources = append(sources, sourceFromDocument(doc))
	}

	return sources
}

func sourceFromDocument(doc *schema.Document) Source {
	if doc == nil {
		return Source{}
	}

	title := metadataString(doc.MetaData, MetaKeyTitle)
	if title == "" {
		title = doc.ID
	}

	return Source{
		ID:     doc.ID,
		Title:  title,
		Source: metadataString(doc.MetaData, MetaKeySource),
		Score:  doc.Score(),
	}
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	if value, ok := metadata[key].(string); ok {
		return value
	}
	return ""
}

func cloneDocuments(docs []*schema.Document) []*schema.Document {
	if docs == nil {
		return nil
	}

	copied := make([]*schema.Document, 0, len(docs))
	for _, doc := range docs {
		copied = append(copied, cloneDocument(doc))
	}

	return copied
}

func cloneDocument(doc *schema.Document) *schema.Document {
	if doc == nil {
		return nil
	}

	copied := &schema.Document{
		ID:      doc.ID,
		Content: doc.Content,
	}
	if doc.MetaData != nil {
		copied.MetaData = make(map[string]any, len(doc.MetaData))
		for key, value := range doc.MetaData {
			copied.MetaData[key] = value
		}
	}
	if score := doc.Score(); score != 0 {
		copied.WithScore(score)
	}

	return copied
}

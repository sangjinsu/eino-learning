package llm

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const DefaultOpenAIModel = "gpt-4.1-mini"

var (
	ErrOpenAIAPIKeyRequired = errors.New("openai config: OPENAI_API_KEY is required")
	ErrOpenAIModelRequired  = errors.New("openai config: OPENAI_MODEL must not be blank")
)

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

func LoadOpenAIConfigFromEnv() OpenAIConfig {
	dotEnv := loadDotEnv()
	model := envValue("OPENAI_MODEL", dotEnv)
	if model == "" {
		model = DefaultOpenAIModel
	}

	return OpenAIConfig{
		APIKey:  envValue("OPENAI_API_KEY", dotEnv),
		Model:   model,
		BaseURL: envValue("OPENAI_BASE_URL", dotEnv),
	}
}

func OpenAIIntegrationEnabled() bool {
	return envValue("RUN_EINO_INTEGRATION", loadDotEnv()) == "1"
}

func (c OpenAIConfig) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return ErrOpenAIAPIKeyRequired
	}
	if strings.TrimSpace(c.Model) == "" {
		return ErrOpenAIModelRequired
	}

	return nil
}

func envValue(key string, dotEnv map[string]string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}

	return strings.TrimSpace(dotEnv[key])
}

func loadDotEnv() map[string]string {
	path, ok := findDotEnv()
	if !ok {
		return nil
	}

	values, err := godotenv.Read(path)
	if err != nil {
		return nil
	}

	return values
}

func findDotEnv() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}

		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return "", false
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

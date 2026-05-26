package openai

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const DefaultModel = "gpt-4.1-mini"

var (
	ErrAPIKeyRequired = errors.New("openai config: OPENAI_API_KEY is required")
	ErrModelRequired  = errors.New("openai config: OPENAI_MODEL must not be blank")
)

type Config struct {
	APIKey  string
	Model   string
	BaseURL string
}

func LoadConfigFromEnv() Config {
	dotEnv := loadDotEnv()
	model := envValue("OPENAI_MODEL", dotEnv)
	if model == "" {
		model = DefaultModel
	}

	return Config{
		APIKey:  envValue("OPENAI_API_KEY", dotEnv),
		Model:   model,
		BaseURL: envValue("OPENAI_BASE_URL", dotEnv),
	}
}

func IntegrationEnabled() bool {
	return envValue("RUN_EINO_INTEGRATION", loadDotEnv()) == "1"
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return ErrAPIKeyRequired
	}
	if strings.TrimSpace(c.Model) == "" {
		return ErrModelRequired
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

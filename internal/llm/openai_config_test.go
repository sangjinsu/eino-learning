package llm

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOpenAIConfigFromEnvUsesSafeDefaults(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("RUN_EINO_INTEGRATION", "")

	got := LoadOpenAIConfigFromEnv()

	if got.APIKey != "" {
		t.Fatal("API key should be empty when OPENAI_API_KEY is not set")
	}
	if got.Model != DefaultOpenAIModel {
		t.Fatalf("model = %q, want %q", got.Model, DefaultOpenAIModel)
	}
	if got.BaseURL != "" {
		t.Fatalf("base URL = %q, want empty", got.BaseURL)
	}
	if OpenAIIntegrationEnabled() {
		t.Fatal("integration should be disabled unless RUN_EINO_INTEGRATION=1")
	}
}

func TestLoadOpenAIConfigFromEnvReadsConfiguredValues(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "gpt-test")
	t.Setenv("OPENAI_BASE_URL", "https://example.test/v1")
	t.Setenv("RUN_EINO_INTEGRATION", "1")

	got := LoadOpenAIConfigFromEnv()

	if got.APIKey != "test-key" {
		t.Fatalf("API key = %q, want configured key", got.APIKey)
	}
	if got.Model != "gpt-test" {
		t.Fatalf("model = %q, want gpt-test", got.Model)
	}
	if got.BaseURL != "https://example.test/v1" {
		t.Fatalf("base URL = %q, want configured URL", got.BaseURL)
	}
	if !OpenAIIntegrationEnabled() {
		t.Fatal("integration should be enabled when RUN_EINO_INTEGRATION=1")
	}
}

func TestLoadOpenAIConfigFromEnvReadsDotEnvFromProjectRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "internal", "llm")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatalf("WriteFile go.mod returned error: %v", err)
	}
	dotEnv := []byte("OPENAI_API_KEY=dotenv-key\nOPENAI_MODEL=gpt-dotenv\nOPENAI_BASE_URL=https://dotenv.test/v1\nRUN_EINO_INTEGRATION=1\n")
	if err := os.WriteFile(filepath.Join(root, ".env"), dotEnv, 0o600); err != nil {
		t.Fatalf("WriteFile .env returned error: %v", err)
	}
	t.Chdir(nested)
	unsetEnv(t, "OPENAI_API_KEY", "OPENAI_MODEL", "OPENAI_BASE_URL", "RUN_EINO_INTEGRATION")

	got := LoadOpenAIConfigFromEnv()

	if got.APIKey != "dotenv-key" {
		t.Fatalf("API key = %q, want dotenv-key", got.APIKey)
	}
	if got.Model != "gpt-dotenv" {
		t.Fatalf("model = %q, want gpt-dotenv", got.Model)
	}
	if got.BaseURL != "https://dotenv.test/v1" {
		t.Fatalf("base URL = %q, want dotenv URL", got.BaseURL)
	}
	if !OpenAIIntegrationEnabled() {
		t.Fatal("integration should be enabled from .env")
	}
}

func TestOpenAIConfigValidateRequiresAPIKeyAndModel(t *testing.T) {
	tests := []struct {
		name string
		cfg  OpenAIConfig
		want error
	}{
		{
			name: "blank API key",
			cfg:  OpenAIConfig{Model: DefaultOpenAIModel},
			want: ErrOpenAIAPIKeyRequired,
		},
		{
			name: "blank model",
			cfg:  OpenAIConfig{APIKey: "test-key"},
			want: ErrOpenAIModelRequired,
		},
		{
			name: "valid config",
			cfg:  OpenAIConfig{APIKey: "test-key", Model: DefaultOpenAIModel},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	previous := make(map[string]string, len(keys))
	existed := make(map[string]bool, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		previous[key] = value
		existed[key] = ok
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) returned error: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if existed[key] {
				_ = os.Setenv(key, previous[key])
				continue
			}
			_ = os.Unsetenv(key)
		}
	})
}

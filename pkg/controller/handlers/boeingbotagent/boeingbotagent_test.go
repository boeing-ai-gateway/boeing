package boeingbotagent

import (
	"context"
	"testing"

	boeingbottypes "github.com/boeing-ai-gateway/boeingbot/pkg/types"
	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	storagescheme "github.com/boeing-ai-gateway/boeing/pkg/storage/scheme"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	sigsyaml "sigs.k8s.io/yaml"
)

func TestChooseModelPrefersKnownNames(t *testing.T) {
	models := []v1.Model{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "ollama-qwen3"},
			Spec: v1.ModelSpec{
				Manifest: types.ModelManifest{
					Name:        "other",
					TargetModel: "some-other-model",
					Active:      true,
					Usage:       types.ModelUsageLLM,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "openai-gpt-5.4"},
			Spec: v1.ModelSpec{
				Manifest: types.ModelManifest{
					Name:        "gpt-5.4",
					TargetModel: "gpt-5.4",
					Active:      true,
					Usage:       types.ModelUsageLLM,
				},
			},
		},
	}

	model, err := chooseModel(context.Background(), nil, "", models, types.DefaultModelAliasTypeLLM)
	if err != nil {
		t.Fatalf("expected model, got error: %v", err)
	}

	if model.Name != "openai-gpt-5.4" {
		t.Fatalf("expected openai-gpt-5.4, got %q", model.Name)
	}
}

func TestChooseModelFallsBackToFirstActiveModel(t *testing.T) {
	models := []v1.Model{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "groq-llama-3.1-70b-versatile"},
			Spec: v1.ModelSpec{
				Manifest: types.ModelManifest{
					Name:        "model-a",
					TargetModel: "model-a",
					Active:      true,
					Usage:       types.ModelUsageLLM,
				},
			},
		},
	}

	model, err := chooseModel(context.Background(), nil, "", models, types.DefaultModelAliasTypeLLM)
	if err != nil {
		t.Fatalf("expected model, got error: %v", err)
	}

	if model.Name != "groq-llama-3.1-70b-versatile" {
		t.Fatalf("expected groq-llama-3.1-70b-versatile, got %q", model.Name)
	}
}

func TestChooseModelPrefersSuggestedOrder(t *testing.T) {
	models := []v1.Model{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "anthropic-claude-sonnet-4-6"},
			Spec: v1.ModelSpec{
				Manifest: types.ModelManifest{
					Name:        "claude-sonnet-4-6",
					TargetModel: "claude-sonnet-4-6",
					Active:      true,
					Usage:       types.ModelUsageLLM,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "openai-gpt-5.4"},
			Spec: v1.ModelSpec{
				Manifest: types.ModelManifest{
					Name:        "gpt-5.4",
					TargetModel: "gpt-5.4",
					Active:      true,
					Usage:       types.ModelUsageLLM,
				},
			},
		},
	}

	model, err := chooseModel(context.Background(), nil, "", models, types.DefaultModelAliasTypeLLM)
	if err != nil {
		t.Fatalf("expected model, got error: %v", err)
	}

	if model.Name != "openai-gpt-5.4" {
		t.Fatalf("expected openai-gpt-5.4, got %q", model.Name)
	}
}

func TestBoeingbotParseModelProviderDeclaredDialectDrivesURL(t *testing.T) {
	h := &Handler{serverURL: "https://boeing.example.com"}

	for _, tc := range []struct {
		dialect     boeingbottypes.Dialect
		wantBaseURL string
	}{
		{boeingbottypes.DialectAnthropicMessages, "https://boeing.example.com/api/llm-proxy/anthropic"},
		{boeingbottypes.DialectOpenAIResponses, "https://boeing.example.com/api/llm-proxy/openai"},
		{boeingbottypes.DialectOpenAIChatCompletions, "https://boeing.example.com/api/llm-proxy"},
		{boeingbottypes.DialectOpenResponses, "https://boeing.example.com/api/llm-proxy"},
		{boeingbottypes.DialectBifrostRequest, "https://boeing.example.com/api/llm-proxy"},
	} {
		model := resolvedLLMModel{
			Name:            "some-model",
			ModelProvider:   "custom-model-provider",
			ProviderDialect: tc.dialect,
		}
		p, _ := h.parseModelProvider(model)
		if p.BaseURL != tc.wantBaseURL {
			t.Errorf("dialect %s: baseURL = %q, want %q", tc.dialect, p.BaseURL, tc.wantBaseURL)
		}
		if p.Dialect != tc.dialect {
			t.Errorf("dialect %s: provider dialect = %q, want same", tc.dialect, p.Dialect)
		}
	}
}

func TestBoeingbotParseModelProviderBuiltinFallbacks(t *testing.T) {
	h := &Handler{serverURL: "https://boeing.example.com"}

	for _, tc := range []struct {
		modelProvider string
		wantDialect   boeingbottypes.Dialect
		wantBaseURL   string
	}{
		{system.OpenAIModelProvider, boeingbottypes.DialectOpenAIResponses, "https://boeing.example.com/api/llm-proxy/openai"},
		{system.AnthropicModelProvider, boeingbottypes.DialectAnthropicMessages, "https://boeing.example.com/api/llm-proxy/anthropic"},
		{"unknown-model-provider", boeingbottypes.DialectOpenResponses, "https://boeing.example.com/api/llm-proxy"},
	} {
		model := resolvedLLMModel{Name: "my-model", ModelProvider: tc.modelProvider}
		p, qualifiedName := h.parseModelProvider(model)
		if p.Dialect != tc.wantDialect {
			t.Errorf("%s: dialect = %q, want %q", tc.modelProvider, p.Dialect, tc.wantDialect)
		}
		if p.BaseURL != tc.wantBaseURL {
			t.Errorf("%s: baseURL = %q, want %q", tc.modelProvider, p.BaseURL, tc.wantBaseURL)
		}
		wantName := tc.modelProvider + "/my-model"
		if qualifiedName != wantName {
			t.Errorf("%s: qualified name = %q, want %q", tc.modelProvider, qualifiedName, wantName)
		}
	}
}

func TestBuildBoeingbotProviderConfigYAMLSingleProvider(t *testing.T) {
	p := boeingbotLLMProvider{
		Name:    "openai-model-provider",
		Dialect: boeingbottypes.DialectOpenAIResponses,
		APIKey:  "${OPENAI_MODEL_PROVIDER_API_KEY}",
		BaseURL: "https://boeing.example.com/api/llm-proxy/openai",
	}

	yaml, err := buildBoeingbotProviderConfigYAML(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg boeingbottypes.Config
	if err := sigsyaml.Unmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("failed to parse output YAML: %v", err)
	}

	if len(cfg.LLMProviders) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(cfg.LLMProviders))
	}
	got := cfg.LLMProviders["openai-model-provider"]
	if got.Dialect != boeingbottypes.DialectOpenAIResponses {
		t.Errorf("dialect = %q, want OpenAIResponses", got.Dialect)
	}
	if got.BaseURL != p.BaseURL {
		t.Errorf("baseURL = %q, want %q", got.BaseURL, p.BaseURL)
	}
}

func TestBuildBoeingbotProviderConfigYAMLMultipleProviders(t *testing.T) {
	openai := boeingbotLLMProvider{
		Name:    "openai-model-provider",
		Dialect: boeingbottypes.DialectOpenAIResponses,
		APIKey:  "${OPENAI_MODEL_PROVIDER_API_KEY}",
		BaseURL: "https://boeing.example.com/api/llm-proxy/openai",
	}
	anthropic := boeingbotLLMProvider{
		Name:    "anthropic-model-provider",
		Dialect: boeingbottypes.DialectAnthropicMessages,
		APIKey:  "${ANTHROPIC_MODEL_PROVIDER_API_KEY}",
		BaseURL: "https://boeing.example.com/api/llm-proxy/anthropic",
	}

	yaml, err := buildBoeingbotProviderConfigYAML(openai, anthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg boeingbottypes.Config
	if err := sigsyaml.Unmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("failed to parse output YAML: %v", err)
	}

	if len(cfg.LLMProviders) != 2 {
		t.Fatalf("expected 2 providers, got %d: %v", len(cfg.LLMProviders), cfg.LLMProviders)
	}
	if cfg.LLMProviders["openai-model-provider"].Dialect != boeingbottypes.DialectOpenAIResponses {
		t.Errorf("openai dialect = %q, want OpenAIResponses", cfg.LLMProviders["openai-model-provider"].Dialect)
	}
	if cfg.LLMProviders["anthropic-model-provider"].Dialect != boeingbottypes.DialectAnthropicMessages {
		t.Errorf("anthropic dialect = %q, want AnthropicMessages", cfg.LLMProviders["anthropic-model-provider"].Dialect)
	}
}

func TestBuildBoeingbotProviderConfigYAMLDeduplicates(t *testing.T) {
	p := boeingbotLLMProvider{
		Name:    "openai-model-provider",
		Dialect: boeingbottypes.DialectOpenAIResponses,
		APIKey:  "${OPENAI_MODEL_PROVIDER_API_KEY}",
		BaseURL: "https://boeing.example.com/api/llm-proxy/openai",
	}

	yaml, err := buildBoeingbotProviderConfigYAML(p, p) // same provider twice
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg boeingbottypes.Config
	if err := sigsyaml.Unmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("failed to parse output YAML: %v", err)
	}

	if len(cfg.LLMProviders) != 1 {
		t.Errorf("expected deduplication to 1 provider, got %d", len(cfg.LLMProviders))
	}
}

func TestResolveModelCarriesProviderAndDialect(t *testing.T) {
	c := fake.NewClientBuilder().
		WithScheme(storagescheme.Scheme).
		WithObjects(
			&v1.DefaultModelAlias{
				TypeMeta:   metav1.TypeMeta{APIVersion: v1.SchemeGroupVersion.String(), Kind: "DefaultModelAlias"},
				ObjectMeta: metav1.ObjectMeta{Name: "llm"},
				Spec: v1.DefaultModelAliasSpec{
					Manifest: types.DefaultModelAliasManifest{Alias: "llm", Model: "groq-llama"},
				},
			},
			&v1.Model{
				TypeMeta:   metav1.TypeMeta{APIVersion: v1.SchemeGroupVersion.String(), Kind: "Model"},
				ObjectMeta: metav1.ObjectMeta{Name: "groq-llama"},
				Spec: v1.ModelSpec{
					Manifest: types.ModelManifest{
						Name:          "groq-llama",
						TargetModel:   "llama-3.1-70b-versatile",
						ModelProvider: "groq-model-provider",
						Active:        true,
						Usage:         types.ModelUsageLLM,
						Dialect:       string(boeingbottypes.DialectOpenAIChatCompletions),
					},
				},
			},
		).Build()

	model, err := resolveModel(context.Background(), c, "", types.DefaultModelAliasTypeLLM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Name != "groq-llama" {
		t.Errorf("Name = %q, want groq-llama", model.Name)
	}
	if model.ModelProvider != "groq-model-provider" {
		t.Errorf("ModelProvider = %q, want groq-model-provider", model.ModelProvider)
	}
	if model.ProviderDialect != boeingbottypes.DialectOpenAIChatCompletions {
		t.Errorf("ProviderDialect = %q, want OpenAIChatCompletions", model.ProviderDialect)
	}
}

// TestMultipleProvidersWhenLLMAndMiniDiffer verifies that when the default LLM and
// mini-LLM models are on different providers, both providers appear in the generated
// boeingbot config YAML.
func TestMultipleProvidersWhenLLMAndMiniDiffer(t *testing.T) {
	h := &Handler{serverURL: "https://boeing.example.com"}

	llmModel := resolvedLLMModel{
		Name:          "anthropic-claude-sonnet-4-6",
		ModelProvider: system.AnthropicModelProvider,
	}
	miniModel := resolvedLLMModel{
		Name:          "openai-gpt-4.1-mini",
		ModelProvider: system.OpenAIModelProvider,
	}

	llmProvider, llmDefault := h.parseModelProvider(llmModel)
	miniProvider, miniDefault := h.parseModelProvider(miniModel)

	if llmDefault != system.AnthropicModelProvider+"/anthropic-claude-sonnet-4-6" {
		t.Errorf("llmDefault = %q, want %s/anthropic-claude-sonnet-4-6", llmDefault, system.AnthropicModelProvider)
	}
	if miniDefault != system.OpenAIModelProvider+"/openai-gpt-4.1-mini" {
		t.Errorf("miniDefault = %q, want %s/openai-gpt-4.1-mini", miniDefault, system.OpenAIModelProvider)
	}

	yaml, err := buildBoeingbotProviderConfigYAML(llmProvider, miniProvider)
	if err != nil {
		t.Fatalf("buildBoeingbotProviderConfigYAML: %v", err)
	}

	var cfg boeingbottypes.Config
	if err := sigsyaml.Unmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("failed to parse output YAML: %v", err)
	}

	if len(cfg.LLMProviders) != 2 {
		t.Fatalf("expected 2 providers (one per model), got %d:\n%s", len(cfg.LLMProviders), yaml)
	}
	if _, ok := cfg.LLMProviders[system.AnthropicModelProvider]; !ok {
		t.Errorf("anthropic-model-provider missing from YAML")
	}
	if _, ok := cfg.LLMProviders[system.OpenAIModelProvider]; !ok {
		t.Errorf("openai-model-provider missing from YAML")
	}
}

func TestChooseModelMiniFallsBackToResolvedLLM(t *testing.T) {
	client := fake.NewClientBuilder().
		WithScheme(storagescheme.Scheme).
		WithObjects(
			&v1.DefaultModelAlias{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "DefaultModelAlias",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "llm",
				},
				Spec: v1.DefaultModelAliasSpec{
					Manifest: types.DefaultModelAliasManifest{
						Alias: "llm",
						Model: "openai-gpt-5.4",
					},
				},
			},
			&v1.Model{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "Model",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "openai-gpt-5.4",
				},
				Spec: v1.ModelSpec{
					Manifest: types.ModelManifest{
						Name:        "gpt-5.4",
						TargetModel: "gpt-5.4",
						Active:      true,
						Usage:       types.ModelUsageLLM,
					},
				},
			},
		).
		Build()

	models := []v1.Model{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "openai-gpt-5.4"},
			Spec: v1.ModelSpec{
				Manifest: types.ModelManifest{
					Name:        "gpt-5.4",
					TargetModel: "gpt-5.4",
					Active:      true,
					Usage:       types.ModelUsageLLM,
				},
			},
		},
	}

	model, err := chooseModel(context.Background(), client, "", models, types.DefaultModelAliasTypeLLMMini)
	if err != nil {
		t.Fatalf("expected model, got error: %v", err)
	}

	if model.Name != "openai-gpt-5.4" {
		t.Fatalf("expected openai-gpt-5.4, got %q", model.Name)
	}
}

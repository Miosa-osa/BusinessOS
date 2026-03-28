package config

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// GetActiveProvider() — provider selection and credential fallback
// ---------------------------------------------------------------------------

func TestGetActiveProvider_OllamaCloudWithKey(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: "sk-test"}
	if got := cfg.GetActiveProvider(); got != "ollama_cloud" {
		t.Errorf("GetActiveProvider() = %q, want %q", got, "ollama_cloud")
	}
}

func TestGetActiveProvider_OllamaCloudWithoutKeyFallsBackToLocal(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: ""}
	if got := cfg.GetActiveProvider(); got != "ollama_local" {
		t.Errorf("GetActiveProvider() = %q, want %q (fallback)", got, "ollama_local")
	}
}

func TestGetActiveProvider_AnthropicWithKey(t *testing.T) {
	cfg := &Config{AIProvider: "anthropic", AnthropicAPIKey: "sk-ant-test"}
	if got := cfg.GetActiveProvider(); got != "anthropic" {
		t.Errorf("GetActiveProvider() = %q, want %q", got, "anthropic")
	}
}

func TestGetActiveProvider_AnthropicWithoutKeyFallsBackToLocal(t *testing.T) {
	cfg := &Config{AIProvider: "anthropic", AnthropicAPIKey: ""}
	if got := cfg.GetActiveProvider(); got != "ollama_local" {
		t.Errorf("GetActiveProvider() = %q, want %q (fallback)", got, "ollama_local")
	}
}

func TestGetActiveProvider_GroqWithKey(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: "gsk-test"}
	if got := cfg.GetActiveProvider(); got != "groq" {
		t.Errorf("GetActiveProvider() = %q, want %q", got, "groq")
	}
}

func TestGetActiveProvider_GroqWithoutKeyFallsBackToLocal(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: ""}
	if got := cfg.GetActiveProvider(); got != "ollama_local" {
		t.Errorf("GetActiveProvider() = %q, want %q (fallback)", got, "ollama_local")
	}
}

func TestGetActiveProvider_OllamaLocalExplicit(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_local"}
	if got := cfg.GetActiveProvider(); got != "ollama_local" {
		t.Errorf("GetActiveProvider() = %q, want %q", got, "ollama_local")
	}
}

func TestGetActiveProvider_UnknownProviderFallsBackToLocal(t *testing.T) {
	cfg := &Config{AIProvider: "unknown_provider"}
	if got := cfg.GetActiveProvider(); got != "ollama_local" {
		t.Errorf("GetActiveProvider() = %q, want %q (fallback for unknown)", got, "ollama_local")
	}
}

func TestGetActiveProvider_EmptyProviderFallsBackToLocal(t *testing.T) {
	cfg := &Config{AIProvider: ""}
	if got := cfg.GetActiveProvider(); got != "ollama_local" {
		t.Errorf("GetActiveProvider() = %q, want %q (fallback for empty)", got, "ollama_local")
	}
}

// ---------------------------------------------------------------------------
// GetActiveModel() — model selection based on active provider
// ---------------------------------------------------------------------------

func TestGetActiveModel_OllamaCloudDefaultModel(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: "sk-test"}
	if got := cfg.GetActiveModel(); got != "llama3.2" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "llama3.2")
	}
}

func TestGetActiveModel_OllamaCloudCustomModel(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: "sk-test", OllamaCloudModel: "mixtral"}
	if got := cfg.GetActiveModel(); got != "mixtral" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "mixtral")
	}
}

func TestGetActiveModel_GroqDefaultModel(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: "gsk-test"}
	if got := cfg.GetActiveModel(); got != "openai/gpt-oss-20b" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "openai/gpt-oss-20b")
	}
}

func TestGetActiveModel_GroqCustomModel(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: "gsk-test", GroqModel: "openai/gpt-oss-20b"}
	if got := cfg.GetActiveModel(); got != "openai/gpt-oss-20b" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "openai/gpt-oss-20b")
	}
}

func TestGetActiveModel_AnthropicDefaultModel(t *testing.T) {
	cfg := &Config{AIProvider: "anthropic", AnthropicAPIKey: "sk-ant-test"}
	if got := cfg.GetActiveModel(); got != "claude-sonnet-4-20250514" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "claude-sonnet-4-20250514")
	}
}

func TestGetActiveModel_AnthropicCustomModel(t *testing.T) {
	cfg := &Config{AIProvider: "anthropic", AnthropicAPIKey: "sk-ant-test", AnthropicModel: "claude-3-opus-20240229"}
	if got := cfg.GetActiveModel(); got != "claude-3-opus-20240229" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "claude-3-opus-20240229")
	}
}

func TestGetActiveModel_OllamaLocalDefaultModel(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_local"}
	if got := cfg.GetActiveModel(); got != "llama3.2:latest" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "llama3.2:latest")
	}
}

func TestGetActiveModel_OllamaLocalCustomModel(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_local", DefaultModel: "mistral:7b"}
	if got := cfg.GetActiveModel(); got != "mistral:7b" {
		t.Errorf("GetActiveModel() = %q, want %q", got, "mistral:7b")
	}
}

func TestGetActiveModel_OllamaCloudWithoutKeyFallsBackToLocalDefault(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: ""}
	// Without key, provider falls back to ollama_local, model should be local default
	if got := cfg.GetActiveModel(); got != "llama3.2:latest" {
		t.Errorf("GetActiveModel() = %q, want %q (local default after fallback)", got, "llama3.2:latest")
	}
}

// ---------------------------------------------------------------------------
// GetModelForProvider() — backwards compatibility delegate
// ---------------------------------------------------------------------------

func TestGetModelForProvider_DelegatesToGetActiveModel(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: "gsk-test", GroqModel: "my-model"}
	if got := cfg.GetModelForProvider(); got != "my-model" {
		t.Errorf("GetModelForProvider() = %q, want %q", got, "my-model")
	}
}

// ---------------------------------------------------------------------------
// Use*() — boolean provider checks
// ---------------------------------------------------------------------------

func TestUseOllamaCloud_ReturnsTrue(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: "key"}
	if !cfg.UseOllamaCloud() {
		t.Error("UseOllamaCloud() = false, want true")
	}
}

func TestUseOllamaCloud_ReturnsFalseWhenFallback(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_cloud", OllamaCloudAPIKey: ""}
	if cfg.UseOllamaCloud() {
		t.Error("UseOllamaCloud() = true, want false (no key)")
	}
}

func TestUseAnthropic_ReturnsTrue(t *testing.T) {
	cfg := &Config{AIProvider: "anthropic", AnthropicAPIKey: "key"}
	if !cfg.UseAnthropic() {
		t.Error("UseAnthropic() = false, want true")
	}
}

func TestUseGroq_ReturnsTrue(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: "key"}
	if !cfg.UseGroq() {
		t.Error("UseGroq() = false, want true")
	}
}

func TestUseOllamaLocal_ReturnsTrue(t *testing.T) {
	cfg := &Config{AIProvider: "ollama_local"}
	if !cfg.UseOllamaLocal() {
		t.Error("UseOllamaLocal() = false, want true")
	}
}

func TestUseOllamaLocal_ReturnsTrueForUnknown(t *testing.T) {
	cfg := &Config{AIProvider: "nonexistent"}
	if !cfg.UseOllamaLocal() {
		t.Error("UseOllamaLocal() = false, want true for unknown provider")
	}
}

// ---------------------------------------------------------------------------
// IsProduction() / LocalModelsAllowed()
// ---------------------------------------------------------------------------

func TestIsProduction_TrueWhenSet(t *testing.T) {
	cfg := &Config{Environment: "production"}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
}

func TestIsProduction_FalseInDevelopment(t *testing.T) {
	cfg := &Config{Environment: "development"}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true, want false")
	}
}

func TestIsProduction_FalseForEmpty(t *testing.T) {
	cfg := &Config{Environment: ""}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true for empty environment, want false")
	}
}

func TestLocalModelsAllowed_TrueInDevelopment(t *testing.T) {
	cfg := &Config{Environment: "development", EnableLocalModels: false}
	if !cfg.LocalModelsAllowed() {
		t.Error("LocalModelsAllowed() = false in development, want true")
	}
}

func TestLocalModelsAllowed_RespectsFlagInProduction(t *testing.T) {
	cfg := &Config{Environment: "production", EnableLocalModels: false}
	if cfg.LocalModelsAllowed() {
		t.Error("LocalModelsAllowed() = true in production with flag=false, want false")
	}
}

func TestLocalModelsAllowed_TrueInProductionWhenEnabled(t *testing.T) {
	cfg := &Config{Environment: "production", EnableLocalModels: true}
	if !cfg.LocalModelsAllowed() {
		t.Error("LocalModelsAllowed() = false in production with flag=true, want true")
	}
}

// ---------------------------------------------------------------------------
// GetSearchProvider() — auto-selection priority
// ---------------------------------------------------------------------------

func TestGetSearchProvider_ExplicitOverride(t *testing.T) {
	cfg := &Config{SearchProvider: "tavily", BraveSearchAPIKey: "brave-key"}
	if got := cfg.GetSearchProvider(); got != "tavily" {
		t.Errorf("GetSearchProvider() = %q, want %q (explicit override)", got, "tavily")
	}
}

func TestGetSearchProvider_AutoSelectsBraveFirst(t *testing.T) {
	cfg := &Config{SearchProvider: "auto", BraveSearchAPIKey: "brave-key", SerperAPIKey: "serper-key"}
	if got := cfg.GetSearchProvider(); got != "brave" {
		t.Errorf("GetSearchProvider() = %q, want %q (brave has priority)", got, "brave")
	}
}

func TestGetSearchProvider_AutoSelectsSerperWhenNoBrave(t *testing.T) {
	cfg := &Config{SearchProvider: "auto", SerperAPIKey: "serper-key"}
	if got := cfg.GetSearchProvider(); got != "serper" {
		t.Errorf("GetSearchProvider() = %q, want %q (serper fallback)", got, "serper")
	}
}

func TestGetSearchProvider_AutoSelectsTavilyWhenNoBraveOrSerper(t *testing.T) {
	cfg := &Config{SearchProvider: "auto", TavilyAPIKey: "tavily-key"}
	if got := cfg.GetSearchProvider(); got != "tavily" {
		t.Errorf("GetSearchProvider() = %q, want %q (tavily fallback)", got, "tavily")
	}
}

func TestGetSearchProvider_AutoFallsBackToDuckDuckGo(t *testing.T) {
	cfg := &Config{SearchProvider: "auto"}
	if got := cfg.GetSearchProvider(); got != "duckduckgo" {
		t.Errorf("GetSearchProvider() = %q, want %q (duckduckgo fallback)", got, "duckduckgo")
	}
}

func TestGetSearchProvider_EmptyStringTriggersAuto(t *testing.T) {
	cfg := &Config{SearchProvider: "", BraveSearchAPIKey: "brave-key"}
	if got := cfg.GetSearchProvider(); got != "brave" {
		t.Errorf("GetSearchProvider() = %q, want %q (empty triggers auto)", got, "brave")
	}
}

// ---------------------------------------------------------------------------
// HasBraveSearch / HasSerper / HasTavily
// ---------------------------------------------------------------------------

func TestHasBraveSearch_TrueWhenSet(t *testing.T) {
	cfg := &Config{BraveSearchAPIKey: "key"}
	if !cfg.HasBraveSearch() {
		t.Error("HasBraveSearch() = false, want true")
	}
}

func TestHasBraveSearch_FalseWhenEmpty(t *testing.T) {
	cfg := &Config{BraveSearchAPIKey: ""}
	if cfg.HasBraveSearch() {
		t.Error("HasBraveSearch() = true, want false")
	}
}

func TestHasSerper_TrueWhenSet(t *testing.T) {
	cfg := &Config{SerperAPIKey: "key"}
	if !cfg.HasSerper() {
		t.Error("HasSerper() = false, want true")
	}
}

func TestHasTavily_TrueWhenSet(t *testing.T) {
	cfg := &Config{TavilyAPIKey: "key"}
	if !cfg.HasTavily() {
		t.Error("HasTavily() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Validate() — security validation
// ---------------------------------------------------------------------------

func TestValidate_EmptySecretKeyFailsInAllEnvironments(t *testing.T) {
	cfg := &Config{Environment: "development", SecretKey: ""}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for empty SECRET_KEY")
	}
	if !strings.Contains(err.Error(), "SECRET_KEY") {
		t.Errorf("error = %q, want to mention SECRET_KEY", err.Error())
	}
}

func TestValidate_InsecureDefaultSecretKeyFails(t *testing.T) {
	cfg := &Config{Environment: "development", SecretKey: "INSECURE-DEFAULT-CHANGE-IN-PRODUCTION"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want error for insecure default SECRET_KEY")
	}
}

func TestValidate_ShortSecretKeyFails(t *testing.T) {
	cfg := &Config{Environment: "development", SecretKey: "too-short"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want error for short SECRET_KEY")
	}
}

func TestValidate_ValidSecretKeyPassesInDevelopment(t *testing.T) {
	key := strings.Repeat("a", 48)
	cfg := &Config{Environment: "development", SecretKey: key}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil in development with valid key", err)
	}
}

func TestValidate_Production_MissingTokenEncryptionKey(t *testing.T) {
	key := strings.Repeat("a", 48)
	cfg := &Config{
		Environment:        "production",
		SecretKey:          key,
		TokenEncryptionKey: "",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for missing TOKEN_ENCRYPTION_KEY")
	}
	if !strings.Contains(err.Error(), "TOKEN_ENCRYPTION_KEY") {
		t.Errorf("error = %q, want to mention TOKEN_ENCRYPTION_KEY", err.Error())
	}
}

func TestValidate_Production_ShortTokenEncryptionKey(t *testing.T) {
	key := strings.Repeat("a", 48)
	cfg := &Config{
		Environment:        "production",
		SecretKey:          key,
		TokenEncryptionKey: "short",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want error for short TOKEN_ENCRYPTION_KEY")
	}
}

func TestValidate_Production_MissingRedisKeyHMACSecret(t *testing.T) {
	key := strings.Repeat("a", 48)
	cfg := &Config{
		Environment:        "production",
		SecretKey:          key,
		TokenEncryptionKey: strings.Repeat("b", 48),
		RedisKeyHMACSecret: "",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for missing REDIS_KEY_HMAC_SECRET")
	}
	if !strings.Contains(err.Error(), "REDIS_KEY_HMAC_SECRET") {
		t.Errorf("error = %q, want to mention REDIS_KEY_HMAC_SECRET", err.Error())
	}
}

func TestValidate_Production_MissingInternalAPISecret(t *testing.T) {
	key := strings.Repeat("a", 48)
	cfg := &Config{
		Environment:        "production",
		SecretKey:          key,
		TokenEncryptionKey: strings.Repeat("b", 48),
		RedisKeyHMACSecret: strings.Repeat("c", 48),
		InternalAPISecret:  "",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for missing INTERNAL_API_SECRET")
	}
	if !strings.Contains(err.Error(), "INTERNAL_API_SECRET") {
		t.Errorf("error = %q, want to mention INTERNAL_API_SECRET", err.Error())
	}
}

func TestValidate_Production_MissingWebhookSigningSecret(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 48),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: "",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for missing WEBHOOK_SIGNING_SECRET")
	}
	if !strings.Contains(err.Error(), "WEBHOOK_SIGNING_SECRET") {
		t.Errorf("error = %q, want to mention WEBHOOK_SIGNING_SECRET", err.Error())
	}
}

func TestValidate_Production_ShortWebhookSigningSecret(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 48),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: "short",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want error for short WEBHOOK_SIGNING_SECRET")
	}
}

func TestValidate_Production_DatabaseURILocalhostFails(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 48),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: strings.Repeat("e", 32),
		DatabaseURL:          "postgres://localhost:5432/db",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for localhost DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "localhost") {
		t.Errorf("error = %q, want to mention localhost", err.Error())
	}
}

func TestValidate_Production_DatabaseURLPlaceholderFails(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 48),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: strings.Repeat("e", 32),
		DatabaseURL:          "postgres://CHANGE_ME@prod/db",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want error for CHANGE_ME placeholder")
	}
}

func TestValidate_Production_EmptyAllowedOriginsFails(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 48),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: strings.Repeat("e", 32),
		DatabaseURL:          "postgres://db.example.com:5432/db",
		AllowedOrigins:       nil,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for empty ALLOWED_ORIGINS")
	}
	if !strings.Contains(err.Error(), "ALLOWED_ORIGINS") {
		t.Errorf("error = %q, want to mention ALLOWED_ORIGINS", err.Error())
	}
}

func TestValidate_Production_WildcardAllowedOriginsFails(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 48),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: strings.Repeat("e", 32),
		DatabaseURL:          "postgres://db.example.com:5432/db",
		AllowedOrigins:       []string{"*"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error for wildcard ALLOWED_ORIGINS")
	}
	if !strings.Contains(err.Error(), "wildcard") {
		t.Errorf("error = %q, want to mention wildcard", err.Error())
	}
}

func TestValidate_Production_AllValidPasses(t *testing.T) {
	cfg := &Config{
		Environment:          "production",
		SecretKey:            strings.Repeat("a", 64),
		TokenEncryptionKey:   strings.Repeat("b", 48),
		RedisKeyHMACSecret:   strings.Repeat("c", 48),
		InternalAPISecret:    strings.Repeat("d", 48),
		WebhookSigningSecret: strings.Repeat("e", 32),
		DatabaseURL:          "postgres://db.example.com:5432/db",
		AllowedOrigins:       []string{"https://app.businessos.com"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil for fully valid production config", err)
	}
}

func TestValidate_MultipleErrorsReported(t *testing.T) {
	cfg := &Config{
		Environment:        "production",
		SecretKey:          "",
		TokenEncryptionKey: "",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want errors for multiple missing secrets")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "SECRET_KEY") {
		t.Error("error should mention SECRET_KEY")
	}
	if !strings.Contains(errMsg, "TOKEN_ENCRYPTION_KEY") {
		t.Error("error should mention TOKEN_ENCRYPTION_KEY")
	}
}

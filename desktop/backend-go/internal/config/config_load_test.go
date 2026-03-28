package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Load() — defaults when no env vars set and no .env file present
// ---------------------------------------------------------------------------

func TestLoad_ReturnsConfigWithoutError(t *testing.T) {
	// Ensure no .env file exists in CWD that could interfere
	// Load() looks for ".env" relative to CWD; our test CWD is the package dir
	// which has no .env, so viper.ReadInConfig() silently fails (desired).
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}

func TestLoad_SetsGlobalAppConfigSingleton(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if AppConfig != cfg {
		t.Error("Load() did not set the global AppConfig singleton")
	}
}

// ---------------------------------------------------------------------------
// Load() — default values when no environment variables are set
// ---------------------------------------------------------------------------

func TestLoad_DefaultEnvironment(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment default = %q, want %q", cfg.Environment, "development")
	}
}

func TestLoad_DefaultServerPort(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerPort != "8001" {
		t.Errorf("ServerPort default = %q, want %q", cfg.ServerPort, "8001")
	}
}

func TestLoad_DefaultBaseURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "http://localhost:8001" {
		t.Errorf("BaseURL default = %q, want %q", cfg.BaseURL, "http://localhost:8001")
	}
}

func TestLoad_DefaultDatabaseURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := "postgres://CHANGE_ME:CHANGE_ME@localhost:5432/business_os"
	if cfg.DatabaseURL != want {
		t.Errorf("DatabaseURL default = %q, want %q", cfg.DatabaseURL, want)
	}
}

func TestLoad_DefaultDatabaseRequired(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.DatabaseRequired {
		t.Error("DatabaseRequired default = false, want true")
	}
}

func TestLoad_DefaultAlgorithm(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Algorithm != "HS256" {
		t.Errorf("Algorithm default = %q, want %q", cfg.Algorithm, "HS256")
	}
}

func TestLoad_DefaultAccessTokenExpireMinutes(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AccessTokenExpireMinutes != 1440 {
		t.Errorf("AccessTokenExpireMinutes default = %d, want %d", cfg.AccessTokenExpireMinutes, 1440)
	}
}

func TestLoad_DefaultAIProvider(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AIProvider != "ollama_cloud" {
		t.Errorf("AIProvider default = %q, want %q", cfg.AIProvider, "ollama_cloud")
	}
}

func TestLoad_DefaultOllamaLocalURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OllamaLocalURL != "http://localhost:11434" {
		t.Errorf("OllamaLocalURL default = %q, want %q", cfg.OllamaLocalURL, "http://localhost:11434")
	}
}

func TestLoad_DefaultOllamaCloudModel(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OllamaCloudModel != "llama3.2" {
		t.Errorf("OllamaCloudModel default = %q, want %q", cfg.OllamaCloudModel, "llama3.2")
	}
}

func TestLoad_DefaultAnthropicModel(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicModel != "claude-sonnet-4-20250514" {
		t.Errorf("AnthropicModel default = %q, want %q", cfg.AnthropicModel, "claude-sonnet-4-20250514")
	}
}

func TestLoad_DefaultGroqModel(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GroqModel != "openai/gpt-oss-20b" {
		t.Errorf("GroqModel default = %q, want %q", cfg.GroqModel, "openai/gpt-oss-20b")
	}
}

func TestLoad_DefaultRedisURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("RedisURL default = %q, want %q", cfg.RedisURL, "redis://localhost:6379/0")
	}
}

func TestLoad_DefaultRedisTLSEnabled(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisTLSEnabled {
		t.Error("RedisTLSEnabled default = true, want false")
	}
}

func TestLoad_DefaultEnableLocalModels(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableLocalModels {
		t.Error("EnableLocalModels default = false, want true")
	}
}

func TestLoad_DefaultPM4PyRustURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PM4PyRustURL != "http://localhost:8090" {
		t.Errorf("PM4PyRustURL default = %q, want %q", cfg.PM4PyRustURL, "http://localhost:8090")
	}
}

func TestLoad_DefaultOSAEnabled(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.OSAEnabled {
		t.Error("OSAEnabled default = false, want true")
	}
}

func TestLoad_DefaultOSATimeout(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OSATimeout != 30 {
		t.Errorf("OSATimeout default = %d, want %d", cfg.OSATimeout, 30)
	}
}

func TestLoad_DefaultOSAMaxRetries(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OSAMaxRetries != 3 {
		t.Errorf("OSAMaxRetries default = %d, want %d", cfg.OSAMaxRetries, 3)
	}
}

func TestLoad_DefaultSearchProvider(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SearchProvider != "auto" {
		t.Errorf("SearchProvider default = %q, want %q", cfg.SearchProvider, "auto")
	}
}

func TestLoad_DefaultSandboxConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "SandboxPortMin", cfg.SandboxPortMin, 9000)
	assertEqual(t, "SandboxPortMax", cfg.SandboxPortMax, 9999)
	assertEqual(t, "SandboxMaxPerUser", cfg.SandboxMaxPerUser, 5)
	assertEqualInt64(t, "SandboxDefaultMemory", cfg.SandboxDefaultMemory, 512*1024*1024)
	assertEqual(t, "SandboxDefaultCPU", cfg.SandboxDefaultCPU, 50000)
	assertEqualInt64(t, "SandboxMaxTotalMemory", cfg.SandboxMaxTotalMemory, 2*1024*1024*1024)
	assertEqualInt64(t, "SandboxMaxTotalStorage", cfg.SandboxMaxTotalStorage, 5*1024*1024*1024)
}

func TestLoad_DefaultBackgroundJobsDisabled(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConversationSummaryJobEnabled {
		t.Error("ConversationSummaryJobEnabled default = true, want false")
	}
	if cfg.BehaviorPatternsJobEnabled {
		t.Error("BehaviorPatternsJobEnabled default = true, want false")
	}
	if cfg.AppProfilerSyncJobEnabled {
		t.Error("AppProfilerSyncJobEnabled default = true, want false")
	}
}

func TestLoad_DefaultConversationSummaryJobConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "ConversationSummaryJobIntervalMinutes", cfg.ConversationSummaryJobIntervalMinutes, 30)
	assertEqual(t, "ConversationSummaryJobBatchSize", cfg.ConversationSummaryJobBatchSize, 25)
	assertEqual(t, "ConversationSummaryJobMaxMessages", cfg.ConversationSummaryJobMaxMessages, 200)
}

func TestLoad_DefaultNATSConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NATSURL != "nats://localhost:4222" {
		t.Errorf("NATSURL default = %q, want %q", cfg.NATSURL, "nats://localhost:4222")
	}
	if cfg.NATSEnabled {
		t.Error("NATSEnabled default = true, want false")
	}
	if cfg.NATSTTL != 24 {
		t.Errorf("NATSTTL default = %d, want %d", cfg.NATSTTL, 24)
	}
}

func TestLoad_DefaultOSAMode(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OSAMode != "local" {
		t.Errorf("OSAMode default = %q, want %q", cfg.OSAMode, "local")
	}
}

func TestLoad_DefaultMIOSACloudURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MIOSACloudURL != "https://api.miosa.ai" {
		t.Errorf("MIOSACloudURL default = %q, want %q", cfg.MIOSACloudURL, "https://api.miosa.ai")
	}
}

// ---------------------------------------------------------------------------
// Load() — environment variable overrides
// ---------------------------------------------------------------------------

func TestLoad_EnvVarOverridesServerPort(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("ServerPort = %q, want %q", cfg.ServerPort, "9090")
	}
}

func TestLoad_EnvVarOverridesAIProvider(t *testing.T) {
	t.Setenv("AI_PROVIDER", "groq")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AIProvider != "groq" {
		t.Errorf("AIProvider = %q, want %q", cfg.AIProvider, "groq")
	}
}

func TestLoad_EnvVarOverridesEnvironment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
	}
}

func TestLoad_EnvVarOverridesSecretKey(t *testing.T) {
	t.Setenv("SECRET_KEY", strings.Repeat("x", 48))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.SecretKey) != 48 {
		t.Errorf("SecretKey length = %d, want 48", len(cfg.SecretKey))
	}
}

func TestLoad_EnvVarOverridesRedisURL(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://redis.prod:6379/1")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisURL != "redis://redis.prod:6379/1" {
		t.Errorf("RedisURL = %q, want %q", cfg.RedisURL, "redis://redis.prod:6379/1")
	}
}

func TestLoad_EnvVarOverridesEnableLocalModels(t *testing.T) {
	t.Setenv("ENABLE_LOCAL_MODELS", "false")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableLocalModels {
		t.Error("EnableLocalModels = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Load() — .env file reading (development only, skipped in production)
// ---------------------------------------------------------------------------

func TestLoad_ProductionMode_IgnoresDotenvFile(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SERVER_PORT", "9999")

	// Create a temporary .env that would set SERVER_PORT to something else
	tmpDir := t.TempDir()
	dotenvPath := filepath.Join(tmpDir, ".env")
	content := "SERVER_PORT=1111\nDATABASE_URL=postgres://override:override@localhost/override\n"
	if err := os.WriteFile(dotenvPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir so Load() finds the .env there
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// In production, .env file is ignored; env var wins
	if cfg.ServerPort != "9999" {
		t.Errorf("SERVER_PORT = %q, want %q (env var should win, .env ignored in production)", cfg.ServerPort, "9999")
	}
}

// ---------------------------------------------------------------------------
// Load() — CORS allowed origins parsing
// ---------------------------------------------------------------------------

func TestLoad_DefaultAllowedOrigins(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// Only http:// and https:// origins pass the filter; app:// is excluded
	expected := []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:3000",
	}
	if len(cfg.AllowedOrigins) != len(expected) {
		t.Fatalf("AllowedOrigins length = %d, want %d (got: %v)", len(cfg.AllowedOrigins), len(expected), cfg.AllowedOrigins)
	}
	for i, want := range expected {
		if cfg.AllowedOrigins[i] != want {
			t.Errorf("AllowedOrigins[%d] = %q, want %q", i, cfg.AllowedOrigins[i], want)
		}
	}
}

func TestLoad_WildcardOriginExcluded(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "*,http://localhost:5173")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			t.Error("AllowedOrigins contains wildcard '*' which should be excluded")
		}
	}
}

func TestLoad_OnlyValidOriginsIncluded(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:5173,notvalid,https://app.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range cfg.AllowedOrigins {
		if o == "notvalid" {
			t.Error("AllowedOrigins contains 'notvalid' which does not start with http:// or https://")
		}
	}
	// Both valid origins should be present
	found5173 := false
	foundExample := false
	for _, o := range cfg.AllowedOrigins {
		if o == "http://localhost:5173" {
			found5173 = true
		}
		if o == "https://app.example.com" {
			foundExample = true
		}
	}
	if !found5173 {
		t.Error("AllowedOrigins missing http://localhost:5173")
	}
	if !foundExample {
		t.Error("AllowedOrigins missing https://app.example.com")
	}
}

// ---------------------------------------------------------------------------
// Load() — DatabaseURL asyncpg format conversion
// ---------------------------------------------------------------------------

func TestLoad_ConvertsAsyncpgURLToPostgres(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql+asyncpg://user:pass@host:5432/db")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cfg.DatabaseURL, "postgres://") {
		t.Errorf("DatabaseURL = %q, want prefix %q", cfg.DatabaseURL, "postgres://")
	}
	if strings.Contains(cfg.DatabaseURL, "asyncpg") {
		t.Error("DatabaseURL still contains 'asyncpg' after conversion")
	}
}

func TestLoad_LeavesStandardPostgresURLUnchanged(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://user:pass@host:5432/db" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://user:pass@host:5432/db")
	}
}

// ---------------------------------------------------------------------------
// Load() — OSA config construction
// ---------------------------------------------------------------------------

func TestLoad_OSAConfigBuiltWhenEnabled(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OSA == nil {
		t.Fatal("OSA config is nil when OSAEnabled is true (default)")
	}
	if cfg.OSA.BaseURL != "http://localhost:8089" {
		t.Errorf("OSA.BaseURL = %q, want %q", cfg.OSA.BaseURL, "http://localhost:8089")
	}
	if cfg.OSA.Timeout != 30*time.Second {
		t.Errorf("OSA.Timeout = %v, want %v", cfg.OSA.Timeout, 30*time.Second)
	}
	if cfg.OSA.MaxRetries != 3 {
		t.Errorf("OSA.MaxRetries = %d, want %d", cfg.OSA.MaxRetries, 3)
	}
	if cfg.OSA.RetryDelay != 2*time.Second {
		t.Errorf("OSA.RetryDelay = %v, want %v", cfg.OSA.RetryDelay, 2*time.Second)
	}
}

func TestLoad_OSAConfigEmptyWhenDisabled(t *testing.T) {
	t.Setenv("OSA_ENABLED", "false")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OSA == nil {
		t.Fatal("OSA config is nil even when disabled")
	}
	if cfg.OSA.BaseURL != "" {
		t.Errorf("OSA.BaseURL = %q, want empty string when disabled", cfg.OSA.BaseURL)
	}
}

// ---------------------------------------------------------------------------
// readDotenvFile() — unit tests with temp files
// ---------------------------------------------------------------------------

func TestReadDotenvFile_ParsesKeyValuePairs(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "KEY1=value1\nKEY2=value2\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if vars["KEY1"] != "value1" {
		t.Errorf("KEY1 = %q, want %q", vars["KEY1"], "value1")
	}
	if vars["KEY2"] != "value2" {
		t.Errorf("KEY2 = %q, want %q", vars["KEY2"], "value2")
	}
}

func TestReadDotenvFile_SkipsComments(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "# This is a comment\nKEY=value\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if _, exists := vars["#"]; exists {
		t.Error("Comment line should not produce a key")
	}
	if vars["KEY"] != "value" {
		t.Errorf("KEY = %q, want %q", vars["KEY"], "value")
	}
}

func TestReadDotenvFile_StripsExportPrefix(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "export DATABASE_URL=postgres://localhost/test\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if vars["DATABASE_URL"] != "postgres://localhost/test" {
		t.Errorf("DATABASE_URL = %q, want %q", vars["DATABASE_URL"], "postgres://localhost/test")
	}
}

func TestReadDotenvFile_StripsDoubleQuotes(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "KEY=\"quoted value\"\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if vars["KEY"] != "quoted value" {
		t.Errorf("KEY = %q, want %q", vars["KEY"], "quoted value")
	}
}

func TestReadDotenvFile_StripsSingleQuotes(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "KEY='single quoted'\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if vars["KEY"] != "single quoted" {
		t.Errorf("KEY = %q, want %q", vars["KEY"], "single quoted")
	}
}

func TestReadDotenvFile_SkipsEmptyLines(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "\n\nKEY=value\n\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if len(vars) != 1 {
		t.Errorf("expected 1 entry, got %d", len(vars))
	}
}

func TestReadDotenvFile_SkipsLinesWithoutEquals(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "INVALID_LINE_NO_EQUALS\nKEY=value\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if _, exists := vars["INVALID_LINE_NO_EQUALS"]; exists {
		t.Error("Line without '=' should not produce a key")
	}
}

func TestReadDotenvFile_ReturnsNilForMissingFile(t *testing.T) {
	vars := readDotenvFile("/nonexistent/path/.env")
	if vars != nil {
		t.Errorf("expected nil for missing file, got %v", vars)
	}
}

func TestReadDotenvFile_TrimsWhitespace(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".env")
	content := "  KEY  =  value  \n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	vars := readDotenvFile(tmpFile)
	if vars["KEY"] != "value" {
		t.Errorf("KEY = %q, want %q (whitespace not trimmed)", vars["KEY"], "value")
	}
}

// ---------------------------------------------------------------------------
// applyDotenvOverrides() — selective field override logic
// ---------------------------------------------------------------------------

func TestApplyDotenvOverrides_SetsEnvironment(t *testing.T) {
	cfg := &Config{Environment: "staging"}
	vars := map[string]string{"ENVIRONMENT": "development"}
	applyDotenvOverrides(cfg, vars)
	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
	}
}

func TestApplyDotenvOverrides_SetsServerPort(t *testing.T) {
	cfg := &Config{ServerPort: "8080"}
	vars := map[string]string{"SERVER_PORT": "3000"}
	applyDotenvOverrides(cfg, vars)
	if cfg.ServerPort != "3000" {
		t.Errorf("ServerPort = %q, want %q", cfg.ServerPort, "3000")
	}
}

func TestApplyDotenvOverrides_SetsDatabaseURL(t *testing.T) {
	cfg := &Config{DatabaseURL: "postgres://old/db"}
	vars := map[string]string{"DATABASE_URL": "postgres://new/db"}
	applyDotenvOverrides(cfg, vars)
	if cfg.DatabaseURL != "postgres://new/db" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://new/db")
	}
}

func TestApplyDotenvOverrides_SetsRedisURL(t *testing.T) {
	cfg := &Config{RedisURL: "redis://old:6379"}
	vars := map[string]string{"REDIS_URL": "redis://new:6379"}
	applyDotenvOverrides(cfg, vars)
	if cfg.RedisURL != "redis://new:6379" {
		t.Errorf("RedisURL = %q, want %q", cfg.RedisURL, "redis://new:6379")
	}
}

func TestApplyDotenvOverrides_DoesNotOverrideUnlistedFields(t *testing.T) {
	cfg := &Config{AIProvider: "groq", GroqAPIKey: "test-key"}
	vars := map[string]string{"SERVER_PORT": "3000"}
	applyDotenvOverrides(cfg, vars)
	if cfg.AIProvider != "groq" {
		t.Errorf("AIProvider = %q, want %q (should not be overridden)", cfg.AIProvider, "groq")
	}
	if cfg.GroqAPIKey != "test-key" {
		t.Errorf("GroqAPIKey = %q, want %q (should not be overridden)", cfg.GroqAPIKey, "test-key")
	}
}

func TestApplyDotenvOverrides_EmptyVarsDoesNotChangeConfig(t *testing.T) {
	cfg := &Config{Environment: "production", ServerPort: "8001"}
	vars := map[string]string{}
	applyDotenvOverrides(cfg, vars)
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
	}
	if cfg.ServerPort != "8001" {
		t.Errorf("ServerPort = %q, want %q", cfg.ServerPort, "8001")
	}
}

func TestApplyDotenvOverrides_DatabaseRequiredTrueFromEnv(t *testing.T) {
	cfg := &Config{DatabaseRequired: false}
	vars := map[string]string{"DATABASE_REQUIRED": "true"}
	applyDotenvOverrides(cfg, vars)
	if !cfg.DatabaseRequired {
		t.Error("DatabaseRequired = false, want true")
	}
}

func TestApplyDotenvOverrides_DatabaseRequiredFalseFromEnv(t *testing.T) {
	cfg := &Config{DatabaseRequired: true}
	vars := map[string]string{"DATABASE_REQUIRED": "false"}
	applyDotenvOverrides(cfg, vars)
	if cfg.DatabaseRequired {
		t.Error("DatabaseRequired = true, want false")
	}
}

func TestApplyDotenvOverrides_DatabaseRequiredNumericOne(t *testing.T) {
	cfg := &Config{DatabaseRequired: false}
	vars := map[string]string{"DATABASE_REQUIRED": "1"}
	applyDotenvOverrides(cfg, vars)
	if !cfg.DatabaseRequired {
		t.Error("DATABASE_REQUIRED='1' should set DatabaseRequired to true")
	}
}

func TestApplyDotenvOverrides_GoogleOAuthFields(t *testing.T) {
	cfg := &Config{}
	vars := map[string]string{
		"GOOGLE_CLIENT_ID":     "client-id",
		"GOOGLE_CLIENT_SECRET": "client-secret",
		"GOOGLE_REDIRECT_URI":  "http://localhost/callback",
	}
	applyDotenvOverrides(cfg, vars)
	if cfg.GoogleClientID != "client-id" {
		t.Errorf("GoogleClientID = %q, want %q", cfg.GoogleClientID, "client-id")
	}
	if cfg.GoogleClientSecret != "client-secret" {
		t.Errorf("GoogleClientSecret = %q, want %q", cfg.GoogleClientSecret, "client-secret")
	}
	if cfg.GoogleRedirectURI != "http://localhost/callback" {
		t.Errorf("GoogleRedirectURI = %q, want %q", cfg.GoogleRedirectURI, "http://localhost/callback")
	}
}

// ---------------------------------------------------------------------------
// Helper assertions
// ---------------------------------------------------------------------------

func assertEqual(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

func assertEqualInt64(t *testing.T, name string, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

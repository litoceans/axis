package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cachepkg "github.com/godlabs/axis/internal/cache"
	"github.com/godlabs/axis/internal/cost"
	"github.com/godlabs/axis/internal/gateway"
	"github.com/godlabs/axis/internal/providers"
	"github.com/godlabs/axis/internal/ratelimit"
	"github.com/godlabs/axis/internal/router"
	"github.com/godlabs/axis/internal/storage"
	"github.com/godlabs/axis/internal/telemetry"
	"github.com/godlabs/axis/pkg/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	// Initialize zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Initialize storage
	storage, err := storage.New(getConfigString("database.url", "sqlite://./axis.db"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize storage")
	}
	defer storage.Close()

	// Initialize rate limiter
	rateLimiter := ratelimit.New(
		getConfigInt("rate_limits.default_rpm", 1000),
		getConfigInt("rate_limits.default_tpm", 10000000),
	)

	// Initialize cost tracker
	costTracker := cost.New()

	// Initialize OpenTelemetry tracer
	otelConfig := telemetry.Config{
		Enabled:     getConfigBool("telemetry.opentelemetry.enabled", false),
		Endpoint:    getConfigString("telemetry.opentelemetry.otlp_endpoint", "http://localhost:4317"),
		ServiceName: getConfigString("telemetry.opentelemetry.service_name", "axis-gateway"),
		SampleRatio: getConfigFloat("telemetry.opentelemetry.sampling_ratio", 1.0),
	}
	tracer, err := telemetry.New(context.Background(), otelConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize tracer")
	}
	defer tracer.Shutdown(context.Background())

	// Initialize health monitor
	healthConfig := router.HealthConfig{
		Window:             getConfigDuration("routing.health.window", 5*time.Minute),
		ErrorThreshold:     getConfigFloat("routing.health.error_threshold", 0.05),
		LatencyThresholdMs: getConfigInt64("routing.health.latency_threshold_ms", 5000),
	}
	healthMonitor := router.NewHealthMonitor(healthConfig)

	// Initialize LRU cache (using defaults)
	lruCache := router.NewCache(true, 100000, 1*time.Hour)

	// Initialize routing config
	routingConfig := types.RoutingConfig{
		DefaultChain:  getConfigString("routing.default_chain", "reliable-balanced"),
		Chains:        getConfigChains(),
		StickySession: getConfigBool("routing.sticky_session", false),
	}

	// Initialize router
	r := router.New(routingConfig, healthMonitor, lruCache, tracer)

	// Register providers
	registerProviders(r)

	// Initialize gateway handler
	handler := gateway.NewHandler(r, storage, rateLimiter, costTracker, tracer)

	// Initialize semantic cache if enabled
	if getConfigBool("cache.semantic.enabled", false) {
		semanticConfig := cachepkg.SemanticConfig{
			Enabled:             true,
			QdrantURL:          getConfigString("cache.semantic.qdrant_url", "localhost:6334"),
			Collection:         getConfigString("cache.semantic.collection", "axis_cache"),
			SimilarityThreshold: getConfigFloat("cache.semantic.similarity_threshold", 0.92),
			TTL:                getConfigDuration("cache.semantic.ttl", 168*time.Hour),
			EmbedderModel:      getConfigString("cache.semantic.embedder_model", "nomic-embed-text"),
			EmbedderProvider:   getConfigString("cache.semantic.embedder_provider", "ollama"),
			VectorSize:         getConfigInt("cache.semantic.vector_size", 768),
		}

		var embedder cachepkg.Embedder
		switch semanticConfig.EmbedderProvider {
		case "ollama":
			embedder = cachepkg.NewOllamaEmbedder(
				os.Getenv("OLLAMA_BASE_URL"),
				semanticConfig.EmbedderModel,
			)
		case "openai":
			embedder = cachepkg.NewOpenAIEmbedder(
				os.Getenv("OPENAI_EMBEDDER_BASE_URL"),
				os.Getenv("OPENAI_API_KEY"),
				semanticConfig.EmbedderModel,
			)
		default:
			embedder = cachepkg.NewOllamaEmbedder(
				os.Getenv("OLLAMA_BASE_URL"),
				semanticConfig.EmbedderModel,
			)
		}

		semanticCache, err := cachepkg.NewSemanticCache(semanticConfig, embedder)
		if err != nil {
			log.Warn().Err(err).Msg("failed to initialize semantic cache")
		} else {
			handler.SetSemanticCache(semanticCache)
			log.Info().Str("collection", semanticConfig.Collection).Msg("semantic cache enabled")
		}
	}

	// Create HTTP server
	server := gateway.NewServer(
		getConfigString("server.host", "0.0.0.0"),
		getConfigInt("server.port", 8080),
		getConfigDuration("server.read_timeout", 60*time.Second),
		getConfigDuration("server.write_timeout", 120*time.Second),
		getConfigDuration("server.idle_timeout", 120*time.Second),
		getConfigInt("server.max_connections", 10000),
		handler,
	)

	// Start Prometheus metrics server (if enabled)
	if getConfigBool("telemetry.prometheus.enabled", true) {
		metricsPort := getConfigInt("telemetry.prometheus.port", 9090)
		go func() {
			mux := http.NewServeMux()
			mux.Handle(getConfigString("telemetry.prometheus.path", "/metrics"), promhttp.Handler())
			addr := fmt.Sprintf(":%d", metricsPort)
			log.Info().Int("port", metricsPort).Msg("starting Prometheus metrics server")
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Error().Err(err).Msg("Prometheus metrics server failed")
			}
		}()
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("host", getConfigString("server.host", "0.0.0.0")).
			Int("port", getConfigInt("server.port", 8080)).
			Msg("starting Axis gateway")
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	// Close router (and providers)
	if err := r.Close(); err != nil {
		log.Error().Err(err).Msg("router close error")
	}

	log.Info().Msg("Axis gateway stopped")
}

func loadConfig() error {
	// Set config file path
	viper.SetConfigName("axis")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/axis/")
	viper.AddConfigPath("$HOME/.axis")

	// Enable environment variable override
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "60s")
	viper.SetDefault("server.write_timeout", "120s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.max_connections", 10000)
	viper.SetDefault("database.url", "sqlite://./axis.db")
	viper.SetDefault("cache.lru.enabled", true)
	viper.SetDefault("cache.lru.max_entries", 100000)
	viper.SetDefault("rate_limits.default_rpm", 1000)
	viper.SetDefault("rate_limits.default_tpm", 10000000)
	viper.SetDefault("routing.default_chain", "reliable-balanced")
	viper.SetDefault("telemetry.prometheus.enabled", true)
	viper.SetDefault("telemetry.prometheus.port", 9090)
	viper.SetDefault("telemetry.prometheus.path", "/metrics")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
		log.Warn().Msg("no config file found, using defaults")
	}

	return nil
}

func getConfigString(key, defaultVal string) string {
	return viper.GetString(key)
}

func getConfigInt(key string, defaultVal int) int {
	return viper.GetInt(key)
}

func getConfigInt64(key string, defaultVal int64) int64 {
	return viper.GetInt64(key)
}

func getConfigFloat(key string, defaultVal float64) float64 {
	return viper.GetFloat64(key)
}

func getConfigBool(key string, defaultVal bool) bool {
	return viper.GetBool(key)
}

func getConfigDuration(key string, defaultVal time.Duration) time.Duration {
	return viper.GetDuration(key)
}

func getConfigChains() map[string][]types.ChainStep {
	chains := make(map[string][]types.ChainStep)

	// Reliable balanced chain
	chains["reliable-balanced"] = []types.ChainStep{
		{Model: "gpt-4o-mini", Provider: "openai", MaxLatencyMs: 3000, MaxRetries: 2, Weight: 1, MaxContextTokens: 128000},
		{Model: "claude-3-5-haiku", Provider: "anthropic", MaxLatencyMs: 4000, MaxRetries: 2, Weight: 1, MaxContextTokens: 200000},
		{Model: "gemini-1.5-flash", Provider: "google", MaxLatencyMs: 5000, MaxRetries: 2, Weight: 1, MaxContextTokens: 1000000},
	}

	// Quality first chain
	chains["quality-first"] = []types.ChainStep{
		{Model: "claude-3-5-sonnet", Provider: "anthropic", MaxLatencyMs: 15000, MaxRetries: 3, MaxContextTokens: 200000},
		{Model: "gpt-4o", Provider: "openai", MaxLatencyMs: 15000, MaxRetries: 3, MaxContextTokens: 128000},
	}

	// Cost conscious chain
	chains["cost-conscious"] = []types.ChainStep{
		{Model: "llama3.3-70b-instruct", Provider: "ollama", MaxLatencyMs: 8000, MaxRetries: 2, MaxContextTokens: 128000},
		{Model: "gpt-4o-mini", Provider: "openai", MaxLatencyMs: 2000, MaxRetries: 2, MaxContextTokens: 128000},
	}

	// Fast local chain
	chains["fast-local"] = []types.ChainStep{
		{Model: "llama3.3", Provider: "ollama", MaxLatencyMs: 5000, MaxRetries: 1, MaxContextTokens: 128000},
		{Model: "qwen2.5-72b-instruct", Provider: "ollama", MaxLatencyMs: 8000, MaxRetries: 1, MaxContextTokens: 128000},
	}

	return chains
}

func registerProviders(r *router.Router) {
	// OpenAI
	openAIProvider := providers.NewOpenAIProvider(
		os.Getenv("OPENAI_API_KEY"),
		viper.GetString("providers.openai.base_url"),
		viper.GetDuration("providers.openai.timeout"),
		viper.GetInt("providers.openai.max_retries"),
	)
	r.RegisterProvider("openai", openAIProvider)

	// Anthropic
	anthropicProvider := providers.NewAnthropicProvider(
		os.Getenv("ANTHROPIC_API_KEY"),
		viper.GetString("providers.anthropic.base_url"),
		viper.GetDuration("providers.anthropic.timeout"),
		viper.GetInt("providers.anthropic.max_retries"),
	)
	r.RegisterProvider("anthropic", anthropicProvider)

	// Google
	googleProvider := providers.NewGoogleProvider(
		os.Getenv("GOOGLE_API_KEY"),
		viper.GetString("providers.google.base_url"),
		viper.GetDuration("providers.google.timeout"),
		viper.GetInt("providers.google.max_retries"),
	)
	r.RegisterProvider("google", googleProvider)

	// Ollama
	ollamaProvider := providers.NewOllamaProvider(
		os.Getenv("OLLAMA_API_KEY"),
		viper.GetString("providers.ollama.base_url"),
		viper.GetDuration("providers.ollama.timeout"),
		viper.GetInt("providers.ollama.max_retries"),
	)
	r.RegisterProvider("ollama", ollamaProvider)
}

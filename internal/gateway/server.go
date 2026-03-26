package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	host       string
	port       int
	handler    *Handler
}

// NewServer creates a new HTTP server
func NewServer(host string, port int, readTimeout, writeTimeout, idleTimeout time.Duration, maxConns int, handler *Handler) *Server {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("POST /v1/chat/completions", handler.HandleChatCompletions)
	mux.HandleFunc("POST /v1/embeddings", handler.HandleEmbeddings)
	mux.HandleFunc("GET /v1/models", handler.HandleListModels)
	mux.HandleFunc("GET /v1/health", handler.HandleHealth)
	mux.HandleFunc("GET /metrics", handler.HandleMetrics)

	// Management API - Keys
	mux.HandleFunc("GET /v1/keys", handler.HandleListKeys)
	mux.HandleFunc("POST /v1/keys", handler.HandleCreateKey)
	mux.HandleFunc("DELETE /v1/keys/{key_id}", handler.HandleDeleteKey)
	mux.HandleFunc("POST /v1/keys/{key_id}/rotate", handler.HandleRotateKey)
	mux.HandleFunc("GET /v1/keys/{key_id}/budget", handler.HandleGetBudget)

	// Management API - Usage & Costs
	mux.HandleFunc("GET /v1/usage", handler.HandleGetUsage)
	mux.HandleFunc("GET /v1/costs", handler.HandleGetCosts)

	// Management API - Routing
	mux.HandleFunc("GET /v1/routing/chains", handler.HandleListChains)
	mux.HandleFunc("POST /v1/routing/chains", handler.HandleCreateChain)

	// Management API - Cache
	mux.HandleFunc("GET /v1/cache/stats", handler.HandleGetCacheStats)
	mux.HandleFunc("DELETE /v1/cache", handler.HandleClearCache)

	// Management API - Budget Alerts
	mux.HandleFunc("GET /v1/alerts", handler.HandleListAlerts)
	mux.HandleFunc("POST /v1/alerts/{alert_id}/acknowledge", handler.HandleAcknowledgeAlert)

	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		Handler:        corsMiddleware(mux),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return &Server{
		httpServer: server,
		host:       host,
		port:       port,
		handler:    handler,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Info().Str("host", s.host).Int("port", s.port).Msg("starting HTTP server")
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Config holds server configuration
type Config struct {
	Host           string
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxConnections int
}

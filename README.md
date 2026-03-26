# Axis LLM Gateway

**Universal AI routing layer. Fast, reliable, observable. Built for production.**

> *"One API. Every model. No surprises."*

Axis is a high-performance LLM gateway written in Go that routes requests across every major AI provider through a single unified OpenAI-compatible API.

## Features

- **🚀 High Performance** - Built in Go for sub-millisecond routing overhead and instant cold starts
- **🔄 Intelligent Fallback** - Automatic failover when providers rate limit or fail
- **💰 Cost Tracking** - Finance-grade token counting and cost estimation per request
- **⚡ Rate Limiting** - Per-key RPM/TPM limits with token bucket algorithm
- **📊 Observability** - Prometheus metrics, structured logging, OpenTelemetry tracing
- **🗄️ Caching** - Built-in LRU cache for non-streaming requests
- **🔐 API Key Management** - Secure key storage with SHA-256 hashing
- **🏥 Health Monitoring** - Real-time provider health scoring with automatic deprioritization

## Quick Start

### Docker

```bash
# Pull and run
docker run -d \
  --name axis \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/axis.yaml:/etc/axis/axis.yaml \
  -e OPENAI_API_KEY=sk-... \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  axis-gateway/axis:latest

# Or build your own
docker build -t axis .
```

### Binary

```bash
# Download release
wget https://releases.axis-gateway.dev/axis_1.0.0_linux_amd64.tar.gz
tar -xzf axis_1.0.0_linux_amd64.tar.gz
chmod +x axis

# Configure
cp axis.yaml.example axis.yaml
nano axis.yaml  # Add your API keys

# Run
./axis serve
```

## Configuration

Copy `axis.yaml.example` to `axis.yaml` and configure:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"

routing:
  default_chain: "reliable-balanced"
  chains:
    reliable-balanced:
      - model: "gpt-4o-mini"
        provider: "openai"
      - model: "claude-3-5-haiku"
        provider: "anthropic"
```

## API Usage

Axis is fully OpenAI-compatible. Just change your base URL:

```bash
# Before (direct to OpenAI)
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'

# After (via Axis)
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $AXIS_API_KEY" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'
```

### Using Routing Chains

Use Axis-specific extensions to select routing chains:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $AXIS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "auto",
    "messages": [{"role":"user","content":"Hello"}],
    "extra_body": {
      "axis_model_hint": "fast-local"
    }
  }'
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Axis Gateway                          │
│                                                              │
│  ┌──────────┐  ┌───────────┐  ┌────────────┐               │
│  │  HTTP    │→ │  Router   │→ │  Provider   │               │
│  │  Server  │  │  Engine   │  │  Pool       │               │
│  │  (Go)    │  │  (fallback│  │  (conn      │               │
│  │          │  │  + cache) │  │  pooling)   │               │
│  └──────────┘  └───────────┘  └────────────┘               │
│       ↑               ↑               ↑                       │
│  ┌──────────┐  ┌───────────┐  ┌────────────┐               │
│  │  Rate    │  │  Cost     │  │  Health     │               │
│  │  Limiter │  │  Tracker  │  │  Monitor    │               │
│  │  (per-key│  │  (async)  │  │  (live)     │               │
│  └──────────┘  └───────────┘  └────────────┘               │
└─────────────────────────────────────────────────────────────┘
         ↓ (async)
┌──────────────────┐     ┌──────────────────┐
│  SQLite / Postgres│    │  Prometheus +    │
│  (usage logs, keys│    │  OpenTelemetry   │
└──────────────────────┘    └──────────────────┘
```

## Supported Providers

| Provider | Chat Models | Embeddings |
|----------|-------------|------------|
| OpenAI | gpt-4o, gpt-4o-mini, gpt-4-turbo, o1-preview, o1-mini | text-embedding-3-small, text-embedding-3-large |
| Anthropic | claude-3-5-sonnet, claude-3-5-haiku, claude-3-opus | - |
| Google | gemini-1.5-pro, gemini-1.5-flash, gemini-2.0-flash | text-embedding-004 |
| Ollama | llama3.3, mistral, qwen2.5, phi4, deepseek-r1 | nomic-embed-text |

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/chat/completions` | Chat completion (streaming + non-streaming) |
| POST | `/v1/embeddings` | Generate embeddings |
| GET | `/v1/models` | List available models |
| GET | `/v1/health` | Provider health status |
| GET | `/metrics` | Prometheus metrics |

## Metrics

Prometheus metrics available at `/metrics`:

```
axis_requests_total{model, provider, status_code}
axis_request_duration_seconds{model, provider, cached}
axis_tokens_total{model, provider, type}
axis_cost_total_usd{model, provider}
axis_cache_hits_total{model}
axis_provider_errors_total{provider, error_type}
axis_provider_health_score{provider}
axis_active_requests{model}
```

## Full Documentation

For complete documentation including:
- Database schema
- API reference
- Deployment options
- Team management
- Budget enforcement

See the [Full Specification](./SPEC.md)

## License

MIT License - see LICENSE file for details.

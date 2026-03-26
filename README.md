# Axis LLM Gateway

**Universal AI routing layer. Fast, reliable, observable. Built for production.**

> *"One API. Every model. No surprises."*

Axis is a high-performance LLM gateway written in Go that routes requests across every major AI provider through a single unified OpenAI-compatible API.

## Quick Start

### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/litoceans/axis/main/scripts/install.sh | bash
```

### Windows

```powershell
powershell -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/litoceans/axis/main/scripts/install-windows.ps1 | iex"
```

### Docker

```bash
docker run -p 8080:8080 -v ./axis-data:/data litoceans/axis:latest
```

Or with Docker Compose:

```bash
docker-compose up -d
```

## Features

- **🚀 High Performance** - Built in Go for sub-millisecond routing overhead and instant cold starts
- **🔄 Intelligent Fallback** - Automatic failover when providers rate limit or fail
- **💰 Cost Tracking** - Finance-grade token counting and cost estimation per request
- **⚡ Rate Limiting** - Per-key RPM/TPM limits with token bucket algorithm
- **📊 Observability** - Prometheus metrics, structured logging, OpenTelemetry tracing
- **🗄️ Caching** - Built-in LRU cache for non-streaming requests
- **🔐 API Key Management** - Secure key storage with SHA-256 hashing
- **🏥 Health Monitoring** - Real-time provider health scoring with automatic deprioritization

## Supported Providers

| Provider | Chat Models | Embeddings |
|----------|-------------|------------|
| **OpenAI** | GPT-5.4, GPT-5, GPT-4.1, o4-mini, o3 | text-embedding-3-small, text-embedding-3-large |
| **Anthropic** | Claude 4.6, Claude 4, Claude 3.7 | - |
| **Google** | Gemini 3.0, Gemini 2.5, Gemini 2.0 | text-embedding-004, text-embedding-005 |
| **Ollama** | Llama 4, Llama 3.3, Qwen 3, DeepSeek V3, Mistral | nomic-embed-text, mxbai-embed-large |
| **MiniMax** | MiniMax M2.7, M2.5, M2, VL 2 | - |
| **Moonshot** | Kimi K2.5, Kimi K2, Kimi VL | - |
| **DeepSeek** | DeepSeek V3.5, V3, R1, Coder V2.5 | - |
| **XAI/Grok** | Grok-4, Grok-3, Grok-2 | - |
| **Mistral** | Mistral Large 2, Small 3, Codestral | - |
| **Groq** | Llama 4, Mixtral, Gemma 2 | - |
| **Cohere** | Command R+, Command A | - |
| **Perplexity** | Sonar Pro, Sonar Small | - |
| **Fireworks** | Hosted Llama, Qwen, Mistral | - |
| **Together AI** | Hosted Llama, Qwen, Mistral | - |
| **Cerebras** | Llama 4 on Cerebras | - |

**150+ models total** across all providers.

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
  google:
    api_key: "${GOOGLE_API_KEY}"
  minimax:
    api_key: "${MINIMAX_API_KEY}"
  moonshot:
    api_key: "${MOONSHOT_API_KEY}"
  deepseek:
    api_key: "${DEEPSEEK_API_KEY}"
  xai:
    api_key: "${XAI_API_KEY}"

routing:
  default_chain: "reliable-balanced"
  chains:
    reliable-balanced:
      - model: "gpt-4.1-mini"
        provider: "openai"
      - model: "claude-3-5-haiku-latest"
        provider: "anthropic"
      - model: "gemini-2.0-flash"
        provider: "google"
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
  -d '{"model":"gpt-4.1-mini","messages":[{"role":"user","content":"Hello"}]}'
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

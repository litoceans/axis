# Axis — LLM Gateway
**Universal AI routing layer. Fast, reliable, observable. Built for production.**

> "The router that just works — without the ops headache."

---

## 1. Concept & Vision

Axis is a high-performance LLM gateway that routes requests across every major AI provider through a single unified OpenAI-compatible API. Built entirely in Go for instant cold starts, zero dependency on PyPI/npm at runtime, and sub-millisecond routing overhead. It is not an MVP — it is production software from day one.

**Tagline:** *"One API. Every model. No surprises."*

**Core promises:**
- Finance teams can trust the cost numbers
- DevOps teams can deploy it in minutes with no mandatory database
- Engineering teams get fallback routing that actually works
- Product teams get observability without a performance tax

---

## 2. Design Language

### Visual Identity
- **Aesthetic:** Developer-tool dark — Linear meets Raycast meets Warp. Professional, data-dense, not flashy.
- **Background:** `#0D0D0F`
- **Surface (cards, panels):** `#17171A`
- **Border:** `#2A2A2F`
- **Text primary:** `#EDEDEF`
- **Text muted:** `#71717A`
- **Accent (primary actions):** `#6366F1` (indigo)
- **Success:** `#22C55E`
- **Warning:** `#F59E0B`
- **Danger:** `#EF4444`
- **Model brand colors:** OpenAI `#10A37F`, Anthropic `#D97706`, Google `#4285F4`, Ollama `#7C3AED`, Cohere `#000000`

### Typography
- **UI:** `Inter` or `Geist` — clean, readable at 11-13px
- **Monospace:** `JetBrains Mono` — API keys, latency numbers, code snippets
- **Scale:** 11px (micro labels), 12px (table cells), 13px (body), 15px (section headers), 20px (page titles), 32px (hero stats)

### Motion
- Transitions: 150ms ease-out (opacity, transform)
- No bounce, no spring — data updates in place, no animation
- Loading states: subtle skeleton shimmer, not spinners
- Charts: real-time streaming updates, no transition on data

---

## 3. Architecture

### 3.1 Core Gateway (Go)

**Binary name:** `axis`
**Config file:** `axis.yaml` (YAML, no code)
**No mandatory database for core routing.**

```
┌─────────────────────────────────────────────────────────┐
│                     Axis Gateway                         │
│                                                          │
│  ┌──────────┐  ┌───────────┐  ┌────────────┐            │
│  │  HTTP    │→ │  Router   │→ │  Provider  │            │
│  │  Server  │  │  Engine   │  │  Pool      │            │
│  │  (Go)    │  │  (fallback│  │  (conn     │            │
│  │          │  │  + cache  │  │  pooling)  │            │
│  └──────────┘  └───────────┘  └────────────┘            │
│       ↑               ↑               ↑                  │
│  ┌──────────┐  ┌───────────┐  ┌────────────┐            │
│  │  Rate    │  │  Cost     │  │  Health    │            │
│  │  Limiter │  │  Tracker  │  │  Monitor   │            │
│  │  (per-key│  │  (async)  │  │  (live)    │            │
│  └──────────┘  └───────────┘  └────────────┘            │
│                                                          │
│  ┌──────────────────────────────────────────┐            │
│  │  Semantic Cache  │  In-memory LRU        │            │
│  │  (optional       │  (built-in)           │            │
│  │   Qdrant)        │                       │            │
│  └──────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────┘
         ↓ (async, non-blocking)
┌──────────────────┐     ┌──────────────────┐
│  SQLite / Postgres │    │  Prometheus +    │
│  (usage logs, keys, │    │  OpenTelemetry   │
│   budgets, orgs)   │    │  (telemetry)      │
└──────────────────────┘    └──────────────────┘
```

### 3.2 Provider Support (Complete List)

**Chat Completion Providers:**
- OpenAI (all models: gpt-4o, gpt-4o-mini, gpt-4-turbo, o1-preview, o1-mini, etc.)
- Anthropic (claude-3-5-sonnet, claude-3-5-haiku, claude-3-opus, claude-sonnet-4, etc.)
- Google AI (gemini-1-5-pro, gemini-1-5-flash, gemini-2.0-flash, gemini-exp-1206, etc.)
- Ollama (any model: llama3.3, mistral, qwen2.5, command-r, phi4, deepseek-r1, etc.)
- Cohere (command-r, command-r-plus, c4ai-aya-expanse)
- Mistral AI (mistral-large, mistral-small, codestral, mixtral-8x7b)
- Perplexity (sonar-small, sonar-large, sonar-pro)
- Groq (llama3-8b, llama3-70b, mixtral-8x7b, gemma2-9b — ultra-low latency)
- AWS Bedrock (Claude via AWS, Titan, Llama via Bedrock, Mistral via Bedrock)
- Azure OpenAI (enterprise OpenAI with Azure AD auth)
- Together AI (hundreds of open-source models: Qwen, Yi, DeepSeek, etc.)
- OpenRouter (100+ models via single upstream — for fallback)
- Replicate (any Replicate-hosted model)
- Fireworks AI (firefunction-v1, llama3.3-70b-instruct, etc.)
- Any OpenAI-compatible API (custom endpoints)

**Embedding Providers:**
- OpenAI (text-embedding-3-small, text-embedding-3-large, text-embedding-ada-002)
- Cohere (embed-english-v3.0, embed-multilingual-v3.0)
- Google (text-embedding-004, embedding-001)
- Ollama (nomic-embed-text, mxbai-embed-large, all-mpnet-base-v2, etc.)
- Mistral (mistral-embed)
- Local models via Ollama compatible API

**Multi-modal:**
- Vision: claude-3-opus, claude-3-sonnet, gpt-4o, gemini-1.5-pro, gemini-1.5-flash
- Audio input: whisper-1, distil-whisper-large-v3.5
- Image generation: DALL-E 3, Stable Diffusion via Replicate

**Future (in scope):**
- Audio output / TTS models
- Video models
- Fine-tuning model management

### 3.3 Request Flow

```
Client
  ↓
Axis Gateway
  ├─ Authenticate (API key validation)
  ├─ Rate limit check (per-key RPM/TPM)
  ├─ Semantic cache lookup (if enabled)
  │    └─ Hit → return cached response
  ├─ Cost estimate (pre-check budget)
  ├─ Route to provider (fallback chain)
  │    ├─ Try model A
  │    ├─ 429 or timeout or error → Try model B
  │    ├─ 429 or timeout or error → Try model C
  │    └─ All fail → return aggregated error
  ├─ Stream response to client
  ├─ Async: log request, update cost, update cache
  └─ Async: emit telemetry (traces, metrics)
```

### 3.4 Configuration Schema (axis.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 60s
  write_timeout: 120s
  idle_timeout: 120s
  max_connections: 10000

database:
  # Optional — gateway works without it
  url: "sqlite://./axis.db"        # SQLite default
  # url: "postgres://user:pass@host:5432/axis"  # PostgreSQL for scale
  max_open_conns: 25
  max_idle_conns: 5

cache:
  # Built-in in-memory LRU (always on, free)
  lru:
    enabled: true
    max_entries: 100000
  # Semantic cache (optional — requires Qdrant or Pgvector)
  semantic:
    enabled: true
    provider: qdrant           # or "pgvector"
    url: "http://localhost:6333"
    collection: "axis_cache"
    similarity_threshold: 0.92
    ttl: 168h                  # 7 days
    max_cost_saving: 0.001     # only cache if saves > $0.001

providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
    organization: "${OPENAI_ORG_ID}"   # optional
    base_url: "https://api.openai.com/v1"
    timeout: 60s
    max_retries: 3
    retry_delay: 500ms
    connection_pool_size: 100

  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    base_url: "https://api.anthropic.com/v1"
    timeout: 120s
    max_retries: 3
    retry_delay: 500ms
    connection_pool_size: 50

  google:
    api_key: "${GOOGLE_API_KEY}"
    base_url: "https://generativelanguage.googleapis.com/v1beta"
    timeout: 60s
    max_retries: 3

  ollama:
    api_key: "ollama"           # not required for local
    base_url: "http://localhost:11434/v1"
    timeout: 300s
    connection_pool_size: 10

  # ... add any provider

routing:
  # Default fallback chain (used when no axis_model_hint specified)
  default_chain: "reliable-balanced"
  
  chains:
    reliable-balanced:
      - model: gpt-4o-mini
        provider: openai
        max_latency_ms: 3000
        max_retries: 2
        weight: 1
      - model: claude-3-5-haiku
        provider: anthropic
        max_latency_ms: 4000
        max_retries: 2
        weight: 1
      - model: gemini-1.5-flash
        provider: google
        max_latency_ms: 5000
        max_retries: 2
        fail_open: true
        weight: 1

    quality-first:
      - model: claude-3-5-sonnet
        provider: anthropic
        max_latency_ms: 15000
        max_retries: 3
      - model: gpt-4o
        provider: openai
        max_latency_ms: 15000
        max_retries: 3

    cost-conscious:
      - model: llama3.3-70b-instruct
        provider: together        # or ollama if local
        max_latency_ms: 8000
        max_retries: 2
      - model: gpt-4o-mini
        provider: openai
        max_latency_ms: 2000
        max_retries: 2

    fast-local:
      - model: llama3.3
        provider: ollama
        max_latency_ms: 5000
        max_retries: 1
      - model: qwen2.5-72b-instruct
        provider: ollama
        max_latency_ms: 8000
        max_retries: 1

  # Health-based routing (automatically avoid degraded providers)
  health:
    enabled: true
    window: 300s               # look back 5 minutes
    error_threshold: 0.05       # >5% error rate → avoid provider
    latency_threshold_ms: 5000  # >5s P99 → deprioritize

  # Latency-aware load balancing
  latency:
    enabled: true
    p99_threshold_ms: 5000
    sample_size: 100

keys:
  # API keys stored in database, but bootstrap key here for first admin
  bootstrap_key: "${AXIS_BOOTSTRAP_KEY}"
  storage: "database"           # "database" or "vault"
  max_per_org: 100

rate_limits:
  default_rpm: 1000
  default_tpm: 10000000         # tokens per minute
  enforce_on_streaming: true
  queue_requests: true          # instead of 429, queue up to 5s
  queue_timeout: 5000ms

costs:
  # Cost per 1M tokens (input, output) — used for estimation
  # Real cost calculated from actual usage
  models:
    gpt-4o-mini:        { input: 0.15,   output: 0.60 }
    gpt-4o:             { input: 2.50,   output: 10.00 }
    claude-3-5-sonnet:   { input: 3.00,   output: 15.00 }
    claude-3-5-haiku:    { input: 0.80,   output: 4.00 }
    gemini-1.5-flash:    { input: 0.075,  output: 0.30 }
    # ... more models

telemetry:
  log_requests: true
  log_responses: false         # NEVER log response content — security
  async: true                  # never block request path
  
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
  
  otel:
    enabled: true
    endpoint: "http://localhost:4317"
    service_name: "axis-gateway"

  tracing:
    sample_rate: 0.1           # 10% of requests — don't overwhelm
    include_request_body: false
    include_response_body: false

alerting:
  slack:
    enabled: false
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#axis-alerts"
    
  pagerduty:
    enabled: false
    routing_key: "${PAGERDUTY_KEY}"

  webhook:
    enabled: false
    url: "${ALERT_WEBHOOK_URL}"
    events:
      - provider_error_rate_above_5pct
      - budget_threshold_exceeded
      - provider_down

auth:
  jwt:
    secret: "${AXIS_JWT_SECRET}"
    expiry: 24h
    issuer: "axis-gateway"
  
  api_key:
    format: "axk_{random_32}"  # axk_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
    hash: "sha256"             # never store plain text

ui:
  enabled: true
  port: 3000
  secret_key: "${AXIS_UI_SECRET}"
  session_cookie: "__axis_session"
  allowed_origins:
    - "https://app.axis-gateway.dev"
    - "http://localhost:3000"
```

---

## 4. Feature Specification

### 4.1 Unified OpenAI-Compatible API

**All endpoints OpenAI-compatible** — swap `base_url` in any SDK.

```bash
# Before (direct to OpenAI)
export OPENAI_API_KEY=sk-...
curl https://api.openai.com/v1/chat/completions ...

# After (via Axis)
export AXIS_API_KEY=axk_live_xxxx
curl https://axis.example.com/v1/chat/completions \
  -H "Authorization: Bearer $AXIS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**Supported endpoints:**
```
POST /v1/chat/completions          — streaming + non-streaming
POST /v1/embeddings                — all embedding models
POST /v1/completions               — legacy completion
GET  /v1/models                   — list available models
GET  /v1/models/:model           — model info + pricing
POST /v1/images/generations       — DALL-E, Stable Diffusion
POST /v1/audio/transcriptions     — Whisper
```

**Axis-specific extensions** (via `axis-` prefixed headers or `extra_body`):
```
axis-model-hint: fast-balanced     # use a routing chain instead of a model
axis-cache: true                  # force cache lookup
axis-cache-similarity: 0.95       # override similarity threshold
axis-no-fallback: false           # disable fallback (use exactly this model)
axis-sticky-session: true          # route same session to same model
axis-session-id: user-123         # session affinity key
```

### 4.2 Intelligent Fallback Routing

**Not basic retry logic — actual intelligent routing.**

**Fallback triggers:**
- HTTP 429 (rate limited)
- HTTP 500/502/503/504 (upstream error)
- Latency exceeded `max_latency_ms`
- Connection timeout
- Provider health score below threshold
- Context window exceeded (for model)

**Sticky sessions:**
- Requests with same `axis-session-id` + same `model_hint` route to same provider
- Reduces variability in chat applications
- Configurable: can be disabled per-request

**Fallback chain execution:**
```
Request arrives
  ↓
Check sticky session → route to last provider
  ↓
Try model A
  ├─ Success → stream response, update health score (+)
  ├─ 429 → mark rate limited for 30s, try model B
  ├─ Timeout → mark slow, try model B
  ├─ 500 → mark errored, try model B
  └─ Context too long → skip to larger model
      ↓
Try model B
  ...
  ↓
All failed → return aggregated error with all attempted models
```

**Health monitoring:**
- Rolling window (5 min default) tracks error rate + latency per provider/model
- Providers below `error_threshold` auto-deprioritized
- Health scores exposed via API: `GET /v1/health`
- Dashboard shows provider status with color coding

### 4.3 Semantic Caching

**First-class native support — not an afterthought.**

```
Request arrives
  ↓
Compute embedding of input text (via configured embedder)
  ↓
Query vector DB for similar embeddings (cosine similarity > threshold)
  ↓
Hit found + TTL valid + cost saving > threshold
  → Return cached response (ultra-low latency, zero provider cost)
  ↓
No hit
  → Route to provider → stream response → store embedding + response
```

**Configuration:**
```yaml
cache:
  semantic:
    enabled: true
    provider: qdrant              # qdrant or pgvector
    collection: "axis_cache"
    similarity_threshold: 0.92     # cosine similarity
    ttl: 168h                      # 7 days
    max_age: 720h                  # max 30 days
    min_cost_saving: 0.001         # only cache if > $0.001 saved
    max_entry_size: 100000         # chars (skip huge prompts)
    embedder:
      model: nomic-embed-text      # via Ollama or OpenAI
      provider: ollama
```

**Cache key:** `sha256(organization_id + model + sha256(normalized_prompt))`
**Normalization:** removes extra whitespace, normalizes JSON key order for deterministic hashing.

**Cache management:**
- `DELETE /v1/cache` — clear all cache
- `DELETE /v1/cache?key_id=xxx` — clear cache for specific key
- Dashboard shows cache hit rate, total savings, top cached prompts

### 4.4 Rate Limiting

**Two-layer limiting:**

**Layer 1 — Gateway-level (per API key):**
- RPM (requests per minute): 1–1,000,000 (configurable)
- TPM (tokens per minute): 1,000–100,000,000 (configurable)
- Hard limit: return 429 immediately
- Soft limit: queue up to `queue_timeout` ms, then 429

**Layer 2 — Provider-level (per provider, automatic):**
- Axis tracks provider rate limits from response headers
- Maintains rolling window per provider
- Prevents hitting provider limits before routing

**Adaptive throttling:**
```yaml
rate_limits:
  queue_requests: true           # queue instead of hard 429
  queue_timeout: 5000ms         # wait up to 5s
  backoff_multiplier: 1.5       # exponential backoff
```

### 4.5 Finance-Grade Cost Tracking

**This is where most routers fail. Axis gets it right.**

**Token counting (critical accuracy):**
- Parse every response for `usage` object
- Handle streaming: aggregate tokens from all chunks
- Handle cached responses: correct `cache_read` / `cache_hit` tokens
- Handle multimodal: count image tokens correctly per provider
- Handle error responses: zero cost for errors

**Provider cost reconciliation:**
- Axis tracks actual spend via provider APIs/billing pages (where available)
- Periodic reconciliation job (daily) compares Axis estimates vs provider invoices
- Drift detection: alert if estimated vs actual differs by > 1%

**Cost data stored:**
```
Per-request record:
  - request_id (UUID)
  - key_id
  - org_id
  - model
  - provider
  - input_tokens
  - output_tokens
  - cached_tokens (if applicable)
  - cost_usd (calculated)
  - latency_ms
  - cached (bool)
  - error (bool)
  - timestamp

Aggregations (computed on read, not stored):
  - Cost per key, per org, per model, per day/hour
  - Cost trend (sparkline data)
  - Budget utilization
  - Cost prediction (linear regression on last 30 days)
```

**Budget enforcement:**
```yaml
budgets:
  per_key:
    monthly_limit_usd: 100.00
    action: alert               # or "block"
  per_org:
    monthly_limit_usd: 1000.00
    action: block               # hard stop when exceeded
  global:
    monthly_limit_usd: 10000.00
    action: alert
```

**Alert thresholds:**
- 50% of budget consumed → Slack notification
- 80% of budget consumed → Slack + email
- 100% of budget consumed → block + Slack

### 4.6 Team & Organization Management

**Hierarchy:** `Organization → Team → Member → API Key`

**Roles:**
- `owner` — full access, billing, can delete org
- `admin` — full access except billing and delete
- `developer` — create/manage keys, view own usage
- `viewer` — read-only access to dashboards

**API Keys:**
- Per-member keys (individual tracking)
- Per-service keys (machine-to-machine)
- Per-project keys (isolation per project)
- Keys have: name, scope (which models), rate limits, budget, expiry date
- Key rotation: instant revoke + regenerate
- Key usage: last used timestamp, total spend, total requests

**Environments:**
- `production`, `staging`, `development` — each with separate keys, limits, alerting

### 4.7 Observability Stack

**Prometheus metrics (always on):**
```
axis_requests_total{model, provider, status_code}
axis_request_duration_seconds{model, provider, cached}
axis_tokens_total{model, provider, type}
axis_cost_total_usd{model, provider}
axis_cache_hits_total{model}
axis_cache_misses_total{model}
axis_provider_errors_total{provider, error_type}
axis_provider_health_score{provider}
axis_active_requests{model}
axis_rate_limit_hits_total{key_id}
axis_budget_utilization{key_id, org_id}
```

**Tracing (OpenTelemetry):**
- Distributed trace per request
- Span per: auth, rate_limit, cache_lookup, provider_call, streaming_chunk
- Trace ID propagated via `X-Trace-ID` header
- Export to Jaeger, Zipkin, Tempo, or any OTLP-compatible backend

**Structured logging:**
```json
{
  "level": "info",
  "ts": "2026-03-25T21:45:00Z",
  "request_id": "req_abc123",
  "method": "POST",
  "path": "/v1/chat/completions",
  "model": "gpt-4o-mini",
  "provider": "openai",
  "status": 200,
  "duration_ms": 342,
  "tokens_in": 45,
  "tokens_out": 128,
  "cost_usd": 0.000882,
  "cached": false,
  "key_id": "key_xyz",
  "org_id": "org_godlabs"
}
```

**Never logged:** API keys (only `key_id`), response content, prompt content.

**Alerting events:**
```
- provider_error_rate_above_5pct
- provider_health_critical
- budget_threshold_80pct
- budget_exceeded
- high_latency_p99_above_10s
- rate_limit_blocked
- cache_hit_rate_below_20pct
- certificate_expiring_soon
```

### 4.8 Security

**API Key security:**
- Keys generated with `SHA-256` hash stored in DB (never plaintext)
- All keys prefixed: `axk_live_` (production) or `axk_test_` (test mode)
- Key last-4 displayed in UI (full key shown only once on creation)
- Instant revocation (DB delete, no grace period)

**Request security:**
- TLS 1.2+ required
- CORS: configurable allowed origins
- Request ID on every response (for audit)
- No response content logged

**Audit log:**
- Every key operation: create, rotate, revoke
- Every org operation: invite, remove, role change
- Every budget change
- Immutable append-only log

**Enterprise security:**
- SAML 2.0 / OIDC SSO
- SCIM provisioning (auto-provision/deprovision users)
- IP allowlist for API keys
- Private networking / VPC peering
- Air-gapped deployment (no external dependencies)
- SOC 2 Type II (roadmap)

---

## 5. API Specification

### 5.1 Authentication

**API Key (header):**
```http
Authorization: Bearer axk_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

**Dashboard (session cookie):**
```http
Cookie: __axis_session=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 5.2 Core Endpoints

```http
POST /v1/chat/completions
Content-Type: application/json
Authorization: Bearer {key}

{
  "model": "gpt-4o",                        // or "axis:fast-balanced"
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false,
  "extra_body": {
    "axis_model_hint": "fast-balanced",      // use routing chain
    "axis_cache": true,                      // semantic cache
    "axis_session_id": "user-123"            // sticky session
  }
}

Response (non-streaming):
{
  "id": "chatcmpl_abc123",
  "object": "chat.completion",
  "created": 1711392000,
  "model": "gpt-4o-mini",                    // actual model used
  "provider": "openai",                       // Axis extension
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 15,
    "total_tokens": 35,
    "cost_usd": 0.000234
  },
  "axis": {
    "cached": false,
    "latency_ms": 342,
    "trace_id": "trace_xyz"
  }
}
```

```http
POST /v1/embeddings
{
  "model": "text-embedding-3-small",
  "input": "The food was delicious"
}

Response:
{
  "object": "list",
  "data": [{
    "index": 0,
    "embedding": [ -0.123, 0.456, ... ],
    "model": "text-embedding-3-small"
  }],
  "usage": {
    "prompt_tokens": 8,
    "cost_usd": 0.000002
  }
}
```

### 5.3 Key Management Endpoints

```http
POST /v1/keys
{
  "name": "Production API Key",
  "org_id": "org_abc",
  "team_id": "team_xyz",          // optional
  "scopes": ["chat", "embeddings"],
  "models": ["gpt-4o-mini", "claude-3-5-haiku"],
  "rpm_limit": 1000,
  "tpm_limit": 10000000,
  "monthly_budget_usd": 100.00,
  "expires_at": "2026-12-31T23:59:59Z",
  "environments": ["production"]
}

GET /v1/keys
GET /v1/keys/:key_id
DELETE /v1/keys/:key_id
POST /v1/keys/:key_id/rotate
```

### 5.4 Usage & Analytics Endpoints

```http
GET /v1/usage?key_id=xxx&from=2026-03-01&to=2026-03-25&granularity=day
{
  "usage": [
    {"date": "2026-03-25", "requests": 15234, "tokens_in": 1234567, "tokens_out": 8901234, "cost_usd": 142.50},
    ...
  ],
  "totals": {
    "requests": 452000,
    "tokens_in": 38000000,
    "tokens_out": 142000000,
    "cost_usd": 3847.23
  }
}

GET /v1/costs?org_id=xxx&by=model          # by=model|key|day|provider
GET /v1/models                               # all available models + pricing
GET /v1/providers/:provider/health           # real-time health scores
GET /v1/routing/chains                      # configured routing chains
POST /v1/routing/chains                     # create custom chain
GET /v1/cache/stats                         # cache hit rate, savings
DELETE /v1/cache                            # clear cache
```

### 5.5 Organization Endpoints

```http
POST /v1/orgs                               # create org
GET  /v1/orgs                               # list orgs (for superadmin)
GET  /v1/orgs/:org_id
PUT  /v1/orgs/:org_id

POST /v1/orgs/:org_id/teams
GET  /v1/orgs/:org_id/teams
DELETE /v1/orgs/:org_id/teams/:team_id

POST /v1/orgs/:org_id/members
GET  /v1/orgs/:org_id/members
PUT  /v1/orgs/:org_id/members/:member_id
DELETE /v1/orgs/:org_id/members/:member_id
```

### 5.6 Error Responses

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": {
    "type": "rate_limit_exceeded",
    "message": "Request rate limit exceeded for key 'key_xyz'. Limit: 1000 RPM, current: 1002",
    "limit": 1000,
    "current": 1002,
    "reset_at": "2026-03-25T21:46:00Z",
    "retry_after_ms": 2340
  }
}

HTTP/1.1 400 Bad Request
{
  "error": {
    "type": "invalid_request",
    "message": "Model 'gpt-5' is not available. Available models: gpt-4o, gpt-4o-mini, ..."
  }
}

HTTP/1.1 400 Bad Request
{
  "error": {
    "type": "budget_exceeded",
    "message": "Monthly budget exceeded for key 'key_xyz'. Limit: $100.00, used: $100.12",
    "budget": 100.00,
    "used": 100.12,
    "reset_at": "2026-04-01T00:00:00Z"
  }
}
```

---

## 6. Dashboard UI

### 6.1 Pages

**Overview (Home)**
- Hero stats: Total spend (this month), Requests (today), Active Keys, Cache savings
- 7-day spend sparkline
- Provider health grid: OpenAI 🟢, Anthropic 🟢, Google 🟡, Ollama 🟢
- Top 5 models by spend (bar chart)
- Recent requests table (last 20, real-time updating)
- Alert banner (if any alerts active)

**Keys**
- Table: Key name, Owner, Team, Created, Last used, This month spend, RPM/TPM, Status
- Create key modal (full configuration)
- Click row → Key detail page:
  - Usage chart (requests + cost over time)
  - Cost breakdown by model
  - Rate limit config
  - Budget config
  - Danger zone: Rotate key, Delete key

**Analytics**
- Time-series charts (zoomable date range):
  - Spend ($/day)
  - Requests (count/day)
  - Latency (P50, P95, P99)
  - Error rate (%)
- Filter bar: Model, Provider, Key, Team, Endpoint
- Comparison mode: Compare two time periods
- Export CSV/JSON
- Cost prediction: "At current pace, March spend will be $X"

**Models**
- Grid of all available models with current pricing
- Provider grouping
- Latency (last 24h P50)
- Usage count
- Cost per day
- Toggle: show only models with active usage

**Routing Chains**
- Visual chain builder (drag-and-drop models)
- Per-model config panel (max latency, retries, fail mode)
- Test panel: paste prompt → see which model handles it + latency + cost
- Chain performance: which models in chain get used most

**Cache**
- Overall hit rate (large number)
- Hit rate over time (line chart)
- Estimated savings ($)
- Top cached prompts (anonymized)
- Cache storage size
- TTL distribution
- Actions: Clear all, Clear by key

**Alerts**
- Active alerts list
- Alert history (last 30 days)
- Configure alert channels (Slack, PagerDuty, webhook)

**Settings**
- Profile: name, email, password
- Organization: name, billing info, subscription
- Team management: invite members, assign roles
- Provider keys: configure your API keys for each provider
- Routing: manage fallback chains
- Security: IP allowlist, API key policies
- Integrations: Slack, PagerDuty, SSO

### 6.2 Visual Style

```
┌─────────────────────────────────────────────────────────────────────┐
│  ┌──────┐                                   [Search] [🔔] [Avatar]   │
│  │ Axis │  Overview  Keys  Analytics  Models  Routing  Cache  Alerts│
├──┴──────┴─────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │ $3,847      │  │ 1.2M        │  │ 142          │  │ 67.3%     │ │
│  │ This month  │  │ Requests    │  │ Active keys  │  │ Cache hit │ │
│  │ ▂▃▅▇▅▃▂▅▇  │  │ +12% vs ↑   │  │              │  │ rate      │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └───────────┘ │
│                                                                      │
│  Provider Health                                    Top Models by $  │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │  OpenAI     🟢  99.8%  P99: 890ms                          │    │
│  │  Anthropic  🟢  99.9%  P99: 1.2s                           │    │
│  │  Google     🟡  98.1%  P99: 3.4s                           │    │
│  │  Ollama     🟢  100%   P99: 240ms                          │    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  Recent Requests                                                  │
│  ┌────────────────────────────────────────────────────────────────┐│
│  │ key_xyz    gpt-4o-mini    200  234ms   $0.00023   12:34:01   ││
│  │ key_abc    claude-3-5...  200  1.1s    $0.00182   12:34:00   ││
│  │ key_def    gemini-1.5...  429  120ms   $0.00000   12:33:58   ││
│  └────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### 6.3 Responsive Strategy
- Desktop-first (primary use case)
- Tablet: collapsible sidebar, tables become scrollable
- Mobile: not a primary target, but functional (basic key viewing, no complex analytics)

---

## 7. Database Schema

### 7.1 SQLite / PostgreSQL (unified schema)

```sql
-- Organizations
CREATE TABLE orgs (
  id TEXT PRIMARY KEY,               -- "org_abc123"
  name TEXT NOT NULL,
  plan TEXT DEFAULT 'free',           -- free, pro, team, enterprise
  monthly_budget_usd REAL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Teams
CREATE TABLE teams (
  id TEXT PRIMARY KEY,
  org_id TEXT REFERENCES orgs(id),
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Members
CREATE TABLE members (
  id TEXT PRIMARY KEY,
  org_id TEXT REFERENCES orgs(id),
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  role TEXT CHECK (role IN ('owner', 'admin', 'developer', 'viewer')),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- API Keys
CREATE TABLE keys (
  id TEXT PRIMARY KEY,               -- "key_abc123" (not the actual key)
  key_hash TEXT UNIQUE NOT NULL,     -- sha256 of the actual key
  key_prefix TEXT NOT NULL,          -- "axk_live_" or "axk_test_"
  key_name TEXT,
  org_id TEXT REFERENCES orgs(id),
  team_id TEXT REFERENCES teams(id),
  member_id TEXT REFERENCES members(id),
  scopes TEXT[],                     -- ['chat', 'embeddings']
  models TEXT[],                    -- allowed models, NULL = all
  rpm_limit INTEGER,
  tpm_limit INTEGER,
  monthly_budget_usd REAL,
  environments TEXT[],               -- ['production']
  expires_at TIMESTAMPTZ,
  last_used_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Usage Logs (partitioned by month for PostgreSQL)
CREATE TABLE usage_logs (
  id TEXT PRIMARY KEY,               -- request UUID
  key_id TEXT REFERENCES keys(id),
  org_id TEXT REFERENCES orgs(id),
  model TEXT NOT NULL,
  provider TEXT NOT NULL,
  endpoint TEXT NOT NULL,            -- chat_completions, embeddings
  input_tokens INTEGER,
  output_tokens INTEGER,
  cached_tokens INTEGER DEFAULT 0,
  cost_usd REAL,
  latency_ms INTEGER,
  status_code INTEGER,
  cached BOOLEAN DEFAULT FALSE,
  error BOOLEAN DEFAULT FALSE,
  error_type TEXT,
  trace_id TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Routing Chains
CREATE TABLE routing_chains (
  id TEXT PRIMARY KEY,
  org_id TEXT REFERENCES orgs(id),   -- NULL = global
  name TEXT NOT NULL,
  chains JSONB NOT NULL,             -- the full chain config
  is_default BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Cache entries (for metadata — actual embeddings in vector DB)
CREATE TABLE cache_entries (
  id TEXT PRIMARY KEY,
  key_id TEXT REFERENCES keys(id),
  org_id TEXT REFERENCES orgs(id),
  model TEXT NOT NULL,
  prompt_hash TEXT NOT NULL,
  response_hash TEXT NOT NULL,      -- dedup
  cost_usd REAL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ
);

-- Audit log
CREATE TABLE audit_log (
  id TEXT PRIMARY KEY,
  org_id TEXT REFERENCES orgs(id),
  actor_id TEXT,                     -- member_id or key_id
  actor_type TEXT,                   -- 'member', 'key', 'system'
  action TEXT NOT NULL,              -- 'key.created', 'key.revoked', etc.
  target_type TEXT,
  target_id TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Budget alerts
CREATE TABLE budget_alerts (
  id TEXT PRIMARY KEY,
  org_id TEXT REFERENCES orgs(id),
  key_id TEXT REFERENCES keys(id),
  threshold_pct REAL NOT NULL,      -- 0.5, 0.8, 1.0
  triggered_at TIMESTAMPTZ DEFAULT NOW(),
  acknowledged_at TIMESTAMPTZ,
  acknowledged_by TEXT REFERENCES members(id)
);

-- Indexes
CREATE INDEX idx_usage_logs_key_id ON usage_logs(key_id);
CREATE INDEX idx_usage_logs_org_id ON usage_logs(org_id);
CREATE INDEX idx_usage_logs_created_at ON usage_logs(created_at);
CREATE INDEX idx_usage_logs_model ON usage_logs(model);
CREATE INDEX idx_keys_org_id ON keys(org_id);
CREATE INDEX idx_cache_entries_key_id ON cache_entries(key_id);
CREATE INDEX idx_audit_log_org_id ON audit_log(org_id);
```

---

## 8. Technical Stack

| Layer | Technology |
|-------|-----------|
| **Gateway core** | Go 1.22+ |
| **HTTP server** | net/http (stdlib) + golang.org/x/net (HTTP/2) |
| **Database** | SQLite (dev/small) / PostgreSQL 16+ (prod) |
| **ORM** | sqlx (raw SQL, no ORM overhead) |
| **Vector DB (cache)** | Qdrant or Pgvector |
| **Config** | viper (YAML) |
| **Logging** | zerolog (structured JSON) |
| **Metrics** | prometheus/client_golang |
| **Tracing** | opentelemetry-go |
| **Auth** | golang-jwt/jwt/v5 |
| **Rate limiting** | Custom token bucket (in-memory + Redis for multi-node) |
| **Connection pooling** | golang.org/x/net/proxy (per-provider pools) |
| **Crypto** | Go stdlib (AES-256 for key encryption at rest) |
| **Build** | goreleaser |
| **Container** | Docker multi-stage build (~20MB image) |
| **Dashboard frontend** | React 18 + Vite + TypeScript |
| **UI components** | Radix UI + TailwindCSS |
| **Charts** | Recharts + Tremor |
| **State management** | Zustand |
| **Routing** | React Router v6 |
| **Deployment** | Docker, Kubernetes (Helm), Railway, Fly.io |

---

## 9. Deployment

### 9.1 Self-Hosted (Single Binary)

```bash
# Download
wget https://releases.axis-gateway.dev/axis_1.0.0_linux_amd64.tar.gz
tar -xzf axis_1.0.0_linux_amd64.tar.gz
chmod +x axis

# Configure
cp axis.yaml.example axis.yaml
nano axis.yaml  # add your provider API keys

# Run
./axis serve

# Docker
docker run -d \
  --name axis \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/axis.yaml:/etc/axis/axis.yaml \
  -v $(pwd)/axis.db:/var/lib/axis/axis.db \
  axis-gateway/axis:1.0.0
```

### 9.2 Self-Hosted (Kubernetes)

```bash
helm repo add axis https://charts.axis-gateway.dev
helm install axis axis/axis \
  --set config.enabled=true \
  --set config.secretName=axis-config \
  --set persistence.enabled=true \
  --set persistence.size=50Gi \
  --set resources.limits.cpu=2 \
  --set resources.limits.memory=2Gi
```

### 9.3 Managed SaaS

```
https://app.axis-gateway.dev
├── Free tier:      100k tokens/mo, 3 keys, no CC
├── Pro:            $29/mo — unlimited keys, 1M tokens/mo
├── Team:           $99/mo — teams, SSO, SLA
└── Enterprise:     Custom — dedicated infra, SCIM, SLA, support
```

### 9.4 System Requirements

| Deployment | CPU | RAM | Disk | Database |
|-----------|-----|-----|------|---------|
| Development | 1 core | 512MB | 1GB | SQLite |
| Production (small) | 2 core | 1GB | 10GB | SQLite or Postgres |
| Production (medium) | 4 core | 4GB | 50GB | PostgreSQL |
| Production (large) | 8+ core | 16GB+ | 100GB+ | PostgreSQL + Qdrant |

---

## 10. Competitive Differentiation

| Feature | Axis | OpenRouter | LiteLLM | Portkey |
|---------|------|------------|---------|---------|
| Cold start | <50ms (Go binary) | N/A (managed) | 3-4 seconds (Python) | N/A (managed) |
| Self-hosted | ✅ Single binary, no DB required | ❌ | ✅ Complex (PG+Redis) | ❌ |
| Fallback routing | Declarative, testable, observable | ❌ | Basic | Limited |
| Semantic caching | Native, built-in LRU + Qdrant | ❌ | DIY | ❌ |
| Cost accuracy | Finance-grade (reconciles with provider) | Approximate | Approximate | Approximate |
| Team management | Full RBAC, hierarchies | ❌ | Limited | ✅ |
| Per-key budgets | Hard block + alerts | ❌ | ❌ | ✅ |
| Supply chain risk | Zero (single binary, reproducible build) | Low | 🔴 CRITICAL (PyPI backdoor) | Low |
| Embeddings native | ✅ | Partial | ✅ | ✅ |
| Multi-modal | ✅ | Partial | ✅ | ✅ |
| Database required | Optional (SQLite-free for core) | N/A | PostgreSQL + Redis (mandatory) | N/A |
| P99 at 500 RPS | <300ms | N/A | 90+ seconds | N/A |
| Observability | Prometheus + OTEL, no performance tax | ❌ | Dashboard tax | ✅ |
| Custom providers | ✅ (YAML config) | ❌ | ✅ (code) | Limited |
| OpenTelemetry | Native | ❌ | Partial | ✅ |
| Alerting | Slack, PagerDuty, webhook | ❌ | ❌ | ✅ |

---

## 11. Roadmap

### Phase 1 — Foundation (T=0)
- [ ] Go gateway: HTTP server, config loading, graceful shutdown
- [ ] Provider clients: OpenAI + Anthropic + Google + Ollama
- [ ] Unified `/v1/chat/completions` + `/v1/embeddings`
- [ ] Basic fallback routing (config-based chains)
- [ ] In-memory rate limiter (per-key RPM/TPM)
- [ ] Structured logging (zerolog → stdout)
- [ ] SQLite persistence (usage logs, keys, orgs)
- [ ] Prometheus metrics endpoint
- [ ] OpenTelemetry tracing (basic)
- [ ] API key authentication
- [ ] Build + release pipeline (goreleaser)

### Phase 2 — Intelligence
- [ ] Health-score routing (rolling error rate + latency per provider)
- [ ] Latency-aware load balancing
- [ ] Sticky sessions
- [ ] In-memory LRU cache
- [ ] Semantic cache (Qdrant integration)
- [ ] Context-length-aware routing (skip models that can't fit)
- [ ] Cost estimation + tracking
- [ ] Budget enforcement (soft alert)

### Phase 3 — Observability
- [ ] Finance-grade cost tracking (streaming token aggregation, reconciliation)
- [ ] Budget alerts (50%, 80%, 100% thresholds)
- [ ] Provider health API endpoint
- [ ] Request tracing UI (link trace IDs to dashboard)
- [ ] Latency histogram (P50/P95/P99 per model)
- [ ] Error rate tracking per provider
- [ ] Cache analytics dashboard

### Phase 4 — Team Management
- [ ] Organization + team hierarchy
- [ ] Member management + RBAC
- [ ] Full API key management UI
- [ ] Per-key rate limits + budgets
- [ ] Audit log (immutable, queryable)
- [ ] Environment separation (dev/staging/prod keys)
- [ ] Key rotation (instant revoke + regenerate)

### Phase 5 — Enterprise
- [ ] PostgreSQL schema + migrations
- [ ] Multi-node deployment (shared rate limit state via Redis)
- [ ] SAML 2.0 / OIDC SSO
- [ ] SCIM provisioning
- [ ] IP allowlist per API key
- [ ] Air-gapped deployment
- [ ] Private networking / VPC peering
- [ ] SOC 2 documentation
- [ ] RBAC audit (who accessed what)

### Phase 6 — Scale
- [ ] Connection pool tuning per provider
- [ ] Request queuing (instead of 429s, queue with backpressure)
- [ ] Batch embedding support
- [ ] Streaming token aggregation optimization
- [ ] Horizontal scaling (stateless gateway nodes)
- [ ] Qdrant cluster support for semantic cache
- [ ] PostgreSQL partitioning (usage_logs by month)

### Phase 7 — Product Expansion
- [ ] Image generation endpoints
- [ ] Audio transcription (Whisper)
- [ ] TTS / audio output
- [ ] Fine-tuning model management
- [ ] A/B testing framework (route % of traffic to different models)
- [ ] Cost-aware routing (automatically choose cheapest model for task)
- [ ] Custom model fine-tuning pipeline (upload data, fine-tune, deploy)

---

## 12. Open Questions (Owner: Soul)

1. **Branding:** "Axis" — confirm no trademark conflict? Available domain?
2. **Managed hosting:** We host it for users (SaaS) or pure self-hosted product?
3. **Provider keys:** Users bring their own API keys (we route, they pay), or we aggregate and mark up?
4. **Initial provider list:** Start with OpenAI + Anthropic + Ollama + Google?
5. **Monetization:** One-time license vs. subscription vs. usage-based?
6. **GitHub repo:** Public or private? Open-source core?
7. **CI/CD:** Which cloud for managed hosting? Railway, Fly.io, AWS?

---

## 13. Success Metrics

- [ ] 100+ stars on GitHub (first month)
- [ ] 10 production self-hosted deployments (first month)
- [ ] <50ms P50 gateway overhead (excluding provider latency)
- [ ] 0 provider outages cause hard failures (fallback works)
- [ ] Cost accuracy within 0.1% of provider invoices
- [ ] Cache hit rate >30% for typical workloads
- [ ] Cold start <100ms on Lambda/Cloud Run

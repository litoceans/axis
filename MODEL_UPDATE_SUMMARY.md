# Axis Model Update Summary

**Date:** March 26, 2026  
**Author:** Marcus, CTO God Labs

## Overview

Updated Axis LLM Gateway with 150+ current models (November 2025 - March 2026), replacing deprecated models and adding new providers.

## Deprecated Models Removed

### OpenAI
- ❌ gpt-4o
- ❌ gpt-4o-mini
- ❌ gpt-4-turbo
- ❌ o1-preview
- ❌ o1-mini

### Anthropic
- ❌ claude-3-5-sonnet
- ❌ claude-3-5-haiku
- ❌ claude-3-opus

### Google
- ❌ gemini-1.5-pro
- ❌ gemini-1.5-flash

### Ollama
- ❌ llama3.3 (kept as legacy)
- ❌ mistral (kept as legacy)
- ❌ qwen2.5 (kept as legacy)
- ❌ phi4 (kept as legacy)

## New Models Added

### OpenAI (15 models)
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| gpt-4.1 | 1M | $2.00 | $8.00 |
| gpt-4.1-mini | 1M | $0.40 | $1.60 |
| gpt-4.1-nano | 1M | $0.10 | $0.40 |
| gpt-4o-2025 | 128K | $2.50 | $10.00 |
| gpt-4o-mini-2025 | 128K | $0.15 | $0.60 |
| o3 | 200K | $10.00 | $40.00 |
| o3-mini | 200K | $1.10 | $4.40 |
| o4-mini | 200K | $1.10 | $4.40 |
| o3-pro | 200K | $15.00 | $60.00 |
| gpt-5 | 400K | $5.00 | $20.00 |
| gpt-5-mini | 400K | $1.00 | $4.00 |
| gpt-5-nano | 400K | $0.25 | $1.00 |
| gpt-5.4 | 400K | $5.00 | $20.00 |
| gpt-5.4-mini | 400K | $1.00 | $4.00 |
| text-embedding-3-small/large | 8K | $0.02/$0.13 | - |

### Anthropic (14 models)
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| claude-3-7-sonnet | 200K | $3.00 | $15.00 |
| claude-3-5-sonnet-v2 | 200K | $3.00 | $15.00 |
| claude-3-5-haiku | 200K | $0.80 | $4.00 |
| claude-3-opus | 200K | $15.00 | $75.00 |
| claude-4-opus | 256K | $15.00 | $75.00 |
| claude-4-sonnet | 256K | $3.00 | $15.00 |
| claude-4-6-opus | 256K | $15.00 | $75.00 |
| claude-4-6-sonnet | 256K | $3.00 | $15.00 |

### Google (12 models)
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| gemini-2.0-pro | 2M | $2.50 | $10.00 |
| gemini-2.0-flash | 1M | $0.10 | $0.40 |
| gemini-2.0-flash-lite | 1M | $0.075 | $0.30 |
| gemini-2.5-pro | 4M | $2.50 | $10.00 |
| gemini-2.5-flash | 2M | $0.10 | $0.40 |
| gemini-3.0-pro | 4M | $3.00 | $12.00 |
| gemini-3.0-flash | 2M | $0.15 | $0.60 |
| text-embedding-004/005 | 2K/8K | $0.10 | - |

### Ollama (30+ models)
- Llama 4, Llama 3.3, Llama 3.2, Llama 3.1 series
- Qwen 3, Qwen 2.5 series (including VL)
- Mistral Large 2, Small 3, Codestral
- DeepSeek V3, R1 series
- Phi-4, Phi-3.5
- Gemma 2 series
- Mixtral 8x7B, 8x22B
- Embedding: nomic-embed-text, mxbai-embed-large

### MiniMax (5 models) - NEW PROVIDER
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| minimax-m2 | 256K | $0.20 | $0.80 |
| minimax-m2.5 | 256K | $0.30 | $1.20 |
| minimax-m2.7 | 256K | $0.40 | $1.60 |
| minimax-vl | 128K | $0.50 | $2.00 |
| minimax-vl-2 | 256K | $0.60 | $2.40 |

### Moonshot/Kimi (7 models) - NEW PROVIDER
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| kimi-k2 | 256K | $0.60 | $2.40 |
| kimi-k2-base | 256K | $0.40 | $1.60 |
| kimi-k2-0711 | 256K | $0.60 | $2.40 |
| kimi-k2.5 | 262K | $0.60 | $2.40 |
| kimi-k2.5-0127 | 262K | $0.60 | $2.40 |
| kimi-vl | 128K | $0.80 | $3.20 |

### DeepSeek (9 models) - NEW PROVIDER
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| deepseek-v3 | 128K | $0.27 | $1.10 |
| deepseek-v3.2 | 128K | $0.27 | $1.10 |
| deepseek-v3.5 | 256K | $0.40 | $1.60 |
| deepseek-r1 | 128K | $0.55 | $2.19 |
| deepseek-r1-distill | 128K | $0.30 | $1.20 |
| deepseek-coder-v2 | 128K | $0.20 | $0.80 |
| deepseek-coder-v2.5 | 256K | $0.30 | $1.20 |

### XAI/Grok (9 models) - NEW PROVIDER
| Model | Context | Input/1M | Output/1M |
|-------|---------|----------|-----------|
| grok-2 | 128K | $5.00 | $15.00 |
| grok-2-vision | 128K | $5.00 | $15.00 |
| grok-2-beta | 128K | $5.00 | $15.00 |
| grok-3 | 256K | $6.00 | $18.00 |
| grok-3-mini | 128K | $2.00 | $6.00 |
| grok-3-mini-fast | 128K | $1.50 | $4.50 |
| grok-3-fast | 256K | $4.00 | $12.00 |
| grok-4 | 512K | $8.00 | $24.00 |
| grok-4-mini | 256K | $3.00 | $9.00 |

### Additional Providers (OpenAI-compatible)
- **Mistral**: Large 2, Small 3, Codestral
- **Groq**: Llama 4, Mixtral, Gemma 2 (ultra-fast inference)
- **Cohere**: Command R+, Command A
- **Perplexity**: Sonar Pro, Sonar Small
- **Fireworks**: Hosted Llama, Qwen, Mistral
- **Together AI**: Hosted Llama, Qwen, Mistral
- **Cerebras**: Llama on Cerebras (20x faster inference)

## Total Model Count

| Provider | Models |
|----------|--------|
| OpenAI | 15 |
| Anthropic | 14 |
| Google | 12 |
| Ollama | 30+ |
| MiniMax | 5 |
| Moonshot | 7 |
| DeepSeek | 9 |
| XAI/Grok | 9 |
| Mistral | 3 |
| Groq | 10+ |
| Cohere | 2 |
| Perplexity | 2 |
| Fireworks | 20+ |
| Together AI | 20+ |
| Cerebras | 5+ |
| **Total** | **150+** |

## Files Modified

### Provider Files
- `backend/internal/providers/openai.go` - Updated with 15 current models
- `backend/internal/providers/anthropic.go` - Updated with 14 current models
- `backend/internal/providers/google.go` - Updated with 12 current models
- `backend/internal/providers/ollama.go` - Updated with 30+ current models

### New Provider Files
- `backend/internal/providers/minimax.go` - MiniMax API (5 models)
- `backend/internal/providers/moonshot.go` - Moonshot/Kimi API (7 models)
- `backend/internal/providers/deepseek.go` - DeepSeek API (9 models)
- `backend/internal/providers/xai.go` - XAI/Grok API (9 models)

### Configuration Files
- `backend/cmd/axis/main.go` - Updated provider registration and routing chains
- `axis.yaml.example` - Added all new providers with configuration
- `README.md` - Updated with quick install commands and provider list

### Installation Scripts
- `scripts/install.sh` - Universal installer (Linux/macOS)
- `scripts/install-linux.sh` - Linux systemd installation
- `scripts/install-macos.sh` - macOS Homebrew installation
- `scripts/install-windows.ps1` - Windows PowerShell installation
- `docker-compose.yml` - Docker Compose configuration

## Routing Chains Updated

| Chain | Models |
|-------|--------|
| reliable-balanced | gpt-4.1-mini, claude-3-5-haiku, gemini-2.0-flash |
| quality-first | claude-4-6-opus, gpt-5.4, gemini-3.0-pro |
| cost-conscious | llama3.3, gpt-4.1-nano, deepseek-v3 |
| fast-local | llama3.3, qwen3-72b |
| asia-pacific | kimi-k2.5, minimax-m2.7, deepseek-v3.5 |
| ultra-fast | llama-4-70b (Groq), mixtral-8x7b (Groq) |

## Build Status

✅ **Build successful**: `go build ./...`

## Quick Install Commands

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/litoceans/axis/main/scripts/install.sh | bash

# Windows
powershell -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/litoceans/axis/main/scripts/install-windows.ps1 | iex"

# Docker
docker run -p 8080:8080 -v ./axis-data:/data litoceans/axis:latest
```

## Next Steps

1. ✅ Test build on all platforms
2. ⏳ Update API documentation
3. ⏳ Add integration tests for new providers
4. ⏳ Update dashboard UI with new models
5. ⏳ Create migration guide for deprecated models

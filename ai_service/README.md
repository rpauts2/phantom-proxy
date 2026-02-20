# 🤖 PhantomProxy AI Service

Enterprise AI service for phishing campaign generation and analysis using LangGraph + Llama-3.1.

## Features

- ✅ **Email Generation** - AI-powered phishing email creation
- ✅ **Content Personalization** - Target-specific content adaptation
- ✅ **Credential Analysis** - Password strength and risk assessment
- ✅ **Report Generation** - Automated security assessment reports
- ✅ **Site Analysis** - Target website vulnerability analysis
- ✅ **Chat Interface** - Interactive AI assistant
- ✅ **LangGraph Workflows** - Multi-step AI pipelines
- ✅ **RAG Support** - Retrieval-Augmented Generation (optional)

## Quick Start

### Docker (Recommended)

```bash
# Start with full stack
docker-compose up -d ai-service ollama

# Check logs
docker-compose logs -f ai-service

# Access API
curl http://localhost:8081/health
```

### Local Development

```bash
# Install dependencies
pip install -r requirements.txt

# Install Ollama (https://ollama.ai)
ollama pull llama3.1:70b

# Run service
python main.py

# Or with uvicorn
uvicorn main:app --reload --port 8081
```

## API Endpoints

### Health Check
```bash
GET /health
```

### Generate Email
```bash
POST /v1/generate/email
Content-Type: application/json

{
  "target_data": {
    "name": "John Doe",
    "company": "Acme Corp",
    "position": "CEO",
    "email": "john@acme.com",
    "interests": ["technology", "finance"]
  },
  "template": "microsoft_login",
  "language": "en",
  "tone": "professional"
}
```

### Personalize Content
```bash
POST /v1/personalize
Content-Type: application/json

{
  "content": "Original email text...",
  "target_profile": {
    "name": "John Doe",
    "company": "Acme Corp",
    "role": "CEO",
    "interests": ["technology"]
  }
}
```

### Analyze Credential
```bash
POST /v1/analyze/credential
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "password123",
  "metadata": {}
}
```

### Generate Report
```bash
POST /v1/report/generate
Content-Type: application/json

{
  "campaign_id": "camp_123",
  "campaign_data": {...},
  "report_type": "executive"
}
```

### Chat
```bash
POST /v1/chat
Content-Type: application/json

{
  "messages": [
    {"role": "user", "content": "How to improve phishing emails?"}
  ],
  "system_prompt": "You are a security testing assistant."
}
```

### Analyze Website
```bash
POST /v1/analyze/site?url=https://target.com
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `ollama` | AI provider (ollama, openai, anthropic) |
| `LLM_MODEL` | `llama3.1:70b` | Model to use |
| `LLM_ENDPOINT` | `http://localhost:11434` | Ollama endpoint |
| `OPENAI_API_KEY` | - | OpenAI API key |
| `OPENAI_MODEL` | `gpt-4` | OpenAI model |
| `MAX_TOKENS` | `4096` | Max generation tokens |
| `TEMPERATURE` | `0.7` | Generation temperature |
| `RAG_ENABLED` | `false` | Enable RAG |

### Example .env

```bash
LLM_PROVIDER=ollama
LLM_MODEL=llama3.1:70b
LLM_ENDPOINT=http://ollama:11434
TEMPERATURE=0.7
MAX_TOKENS=4096
RAG_ENABLED=false
```

## Models

### Supported Models

- **Ollama:**
  - `llama3.1:70b` (recommended)
  - `llama3.1:8b`
  - `mistral:7b`
  - `mixtral:8x7b`

- **OpenAI:**
  - `gpt-4`
  - `gpt-4-turbo`
  - `gpt-3.5-turbo`

- **Anthropic:**
  - `claude-3-opus`
  - `claude-3-sonnet`
  - `claude-3-haiku`

## LangGraph Workflows

The service uses LangGraph for multi-step AI workflows:

```python
# Email Generation Workflow
1. Analyze Target → 2. Generate Draft → 3. Refine Content → 4. Quality Check
```

## Integration

### PhantomProxy Core

The AI service integrates with PhantomProxy core:

```yaml
# config.yaml
ai:
  enabled: true
  endpoint: http://ai-service:8081
  timeout: 60s
```

### Python SDK

```python
from phantom_ai import AIClient

client = AIClient("http://localhost:8081")

# Generate email
email = client.generate_email(
    target_data={"name": "John", "company": "Acme"},
    template="microsoft_login"
)

# Analyze credential
analysis = client.analyze_credential("user@example.com", "password123")
```

## Performance

- **Latency:** ~2-5s for email generation (70B model)
- **Throughput:** ~10 requests/minute
- **Memory:** 2GB (service) + 40GB (70B model)

## Troubleshooting

### Ollama Connection Error

```bash
# Check Ollama is running
docker-compose ps ollama

# Pull model
docker-compose exec ollama ollama pull llama3.1:70b
```

### Out of Memory

```bash
# Use smaller model
LLM_MODEL=llama3.1:8b
```

### Slow Generation

```bash
# Use GPU acceleration
# Ensure NVIDIA GPU and docker-compose.yml has GPU config
```

## License

Proprietary - PhantomSec Labs

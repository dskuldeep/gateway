# LLM Gateway

A high-performance, low-latency LLM gateway in Go supporting multiple providers, organizations, projects, and an optional intelligence layer.

## Features

- Multi-provider LLM support (OpenAI, Anthropic, Mistral, Groq, Google AI Studio)
- Organization and project management
- API key management per project and provider
- Optional intelligence layer for response evaluation
- Prometheus metrics for monitoring
- Clerk.dev integration for authentication
- Docker support for easy deployment

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Clerk.dev account
- LLM provider API keys

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/kuldeeppaul/gateway.git
   cd gateway
   ```

2. Create a `.env` file:
   ```bash
   cp .env.example .env
   ```

3. Update the `.env` file with your API keys and configuration:
   ```env
   # Clerk.dev configuration
   CLERK_SECRET_KEY=your_clerk_secret_key

   # LLM Provider API Keys
   OPENAI_API_KEY=your_openai_api_key
   ANTHROPIC_API_KEY=your_anthropic_api_key
   MISTRAL_API_KEY=your_mistral_api_key
   GROQ_API_KEY=your_groq_api_key
   GOOGLE_AI_API_KEY=your_google_ai_api_key

   # Database configuration
   DB_HOST=postgres
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=gateway
   DB_SSL_MODE=disable

   # Server configuration
   SERVER_PORT=8080
   SERVER_READ_TIMEOUT=10s
   SERVER_WRITE_TIMEOUT=10s
   SERVER_IDLE_TIMEOUT=120s

   # Intelligence layer configuration
   INTELLIGENCE_ENABLED=true
   INTELLIGENCE_SAMPLING_RATE=0.1
   INTELLIGENCE_KEYWORDS=safety,relevance,completeness
   INTELLIGENCE_MIN_SCORE=0.7
   INTELLIGENCE_MAX_LATENCY=5s

   # Logging configuration
   LOG_LEVEL=info
   LOG_FORMAT=json
   LOG_OUTPUT=stdout

   # Metrics configuration
   PROMETHEUS_ENABLED=true
   PROMETHEUS_PATH=/metrics
   ```

4. Start the services:
   ```bash
   docker-compose up -d
   ```

## API Endpoints

### LLM Endpoints

- `POST /api/v1/llm/query` - Query an LLM provider
- `GET /api/v1/llm/models` - List available models

### Organization Endpoints

- `GET /api/v1/orgs` - List organizations
- `POST /api/v1/orgs` - Create an organization
- `GET /api/v1/orgs/{id}` - Get organization details
- `PUT /api/v1/orgs/{id}` - Update organization
- `DELETE /api/v1/orgs/{id}` - Delete organization

### Project Endpoints

- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create a project
- `GET /api/v1/projects/{id}` - Get project details
- `PUT /api/v1/projects/{id}` - Update project
- `DELETE /api/v1/projects/{id}` - Delete project

### API Key Endpoints

- `GET /api/v1/api-keys` - List API keys
- `POST /api/v1/api-keys` - Create an API key
- `DELETE /api/v1/api-keys` - Delete an API key

### Health and Metrics

- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint

## Example Usage

### Query an LLM

```bash
curl -X POST http://localhost:8080/api/v1/llm/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_clerk_token" \
  -d '{
    "provider": "groq",
    "model": "mixtral-8x7b-32768",
    "prompt": "What is the capital of France?",
    "max_tokens": 100,
    "temperature": 0.7
  }'
```

### Create an Organization

```bash
curl -X POST http://localhost:8080/api/v1/orgs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_clerk_token" \
  -d '{
    "name": "My Organization"
  }'
```

### Create a Project

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_clerk_token" \
  -d '{
    "organization_id": "org_123",
    "name": "My Project"
  }'
```

### Add an API Key

```bash
curl -X POST http://localhost:8080/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_clerk_token" \
  -d '{
    "organization_id": "org_123",
    "project_id": "proj_456",
    "provider": "groq",
    "api_key": "your_groq_api_key"
  }'
```

## Monitoring

The application exposes Prometheus metrics at `/metrics`. You can access the Prometheus UI at `http://localhost:9090` to view metrics and create dashboards.

## Development

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run tests:
   ```bash
   go test ./...
   ```

3. Build the application:
   ```bash
   go build -o gateway
   ```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
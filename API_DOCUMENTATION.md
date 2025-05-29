# LLM Gateway API Documentation

## Overview

The LLM Gateway is a high-performance, multi-provider LLM service built in Go with comprehensive organization management, project isolation, and monitoring capabilities. It supports multiple LLM providers with role-based access control and real-time metrics.

## Base URL
```
http://localhost:8080
```

## Authentication

All API endpoints (except health and metrics) require authentication using Clerk.dev JWT tokens.

### Authentication Header
```http
Authorization: Bearer <clerk_jwt_token>
```

### Token Claims
The JWT token contains the following claims that are automatically extracted:
- `sub` (subject): User ID
- `org_id`: Organization ID
- `project_id`: Project ID

## User Roles & Access Levels

### 1. Superadmin Level
- **Full system access**: Can manage all organizations, projects, and users
- **Global metrics access**: Can view system-wide performance metrics
- **Provider management**: Can configure LLM provider settings
- **System monitoring**: Access to all Prometheus metrics and dashboards

### 2. Organization Admin Level  
- **Organization management**: Full control over their organization
- **Project management**: Create, update, delete projects within organization
- **Member management**: Add/remove organization members
- **API key management**: Manage API keys for all projects in organization
- **Organization metrics**: Access to organization-scoped metrics

### 3. Project User Level
- **Project access**: Access to assigned projects only
- **LLM queries**: Can make LLM requests within authorized projects
- **Usage metrics**: View usage metrics for accessible projects
- **Limited API key access**: View API keys for authorized projects

---

## API Endpoints

### System Health & Monitoring

#### Health Check
```http
GET /health
```

**Description**: System health status endpoint  
**Authentication**: None required  
**Response**:
```json
{
  "status": "ok"
}
```

#### Prometheus Metrics
```http
GET /metrics
```

**Description**: Prometheus metrics endpoint for monitoring  
**Authentication**: None required  
**Content-Type**: `text/plain`  
**Response**: Prometheus metrics in text format

---

### LLM Service Endpoints

#### Query LLM Provider
```http
POST /v1/llm/query
```

**Description**: Send a query to an LLM provider  
**Authentication**: Required  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "provider": "openai|groq|google|anthropic|mistral",
  "model": "string",
  "prompt": "string",
  "max_tokens": 100,
  "temperature": 0.7,
  "stop": ["string"],
  "stream": false,
  "metadata": {
    "key": "value"
  }
}
```

**Response** (200 OK):
```json
{
  "id": "response_id",
  "provider": "openai",
  "model": "gpt-4",
  "text": "Generated response text",
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 25,
    "total_tokens": 35
  },
  "finish_reason": "stop",
  "metadata": {
    "project_id": "proj_123",
    "user_id": "user_456"
  },
  "latency": "250ms"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request format or unsupported provider/model
- `401 Unauthorized`: Missing or invalid authentication
- `500 Internal Server Error`: LLM provider error or system failure

#### List Available Models
```http
GET /v1/llm/models
```

**Description**: Get list of all available LLM models across providers  
**Authentication**: Required

**Response** (200 OK):
```json
[
  {
    "id": "gpt-4",
    "provider": "openai",
    "name": "GPT-4",
    "description": "OpenAI's most capable model",
    "max_tokens": 8192,
    "cost_per_1k": 0.03
  },
  {
    "id": "mixtral-8x7b-32768",
    "provider": "groq",
    "name": "Mixtral 8x7B",
    "description": "Fast inference model by Mistral",
    "max_tokens": 32768,
    "cost_per_1k": 0.0005
  }
]
```

---

### Organization Management

#### List Organizations
```http
GET /v1/orgs
```

**Description**: List all organizations (filtered by user access level)  
**Authentication**: Required

**Response** (200 OK):
```json
[
  {
    "id": "org_123",
    "name": "Acme Corporation",
    "created_at": "2025-05-29T10:00:00Z",
    "updated_at": "2025-05-29T10:00:00Z"
  }
]
```

#### Create Organization
```http
POST /v1/orgs
```

**Description**: Create a new organization (superadmin only)  
**Authentication**: Required  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "name": "Organization Name"
}
```

**Response** (201 Created):
```json
{
  "id": "org_456",
  "name": "Organization Name",
  "created_at": "2025-05-29T12:00:00Z",
  "updated_at": "2025-05-29T12:00:00Z"
}
```

#### Get Organization
```http
GET /v1/orgs/{id}
```

**Description**: Get organization details by ID  
**Authentication**: Required

**Response** (200 OK):
```json
{
  "id": "org_123",
  "name": "Acme Corporation",
  "created_at": "2025-05-29T10:00:00Z",
  "updated_at": "2025-05-29T10:00:00Z"
}
```

#### Update Organization
```http
PUT /v1/orgs/{id}
```

**Description**: Update organization details (org admin or superadmin)  
**Authentication**: Required  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "name": "Updated Organization Name"
}
```

**Response** (200 OK):
```json
{
  "id": "org_123",
  "name": "Updated Organization Name",
  "created_at": "2025-05-29T10:00:00Z",
  "updated_at": "2025-05-29T12:30:00Z"
}
```

#### Delete Organization
```http
DELETE /v1/orgs/{id}
```

**Description**: Delete organization (superadmin only)  
**Authentication**: Required

**Response**: `204 No Content`

---

### Project Management

#### List Projects
```http
GET /v1/projects
```

**Description**: List projects (filtered by user access)  
**Authentication**: Required

**Response** (200 OK):
```json
[
  {
    "id": "proj_123",
    "organization_id": "org_123",
    "name": "AI Chatbot Project",
    "created_at": "2025-05-29T10:00:00Z",
    "updated_at": "2025-05-29T10:00:00Z"
  }
]
```

#### Create Project
```http
POST /v1/projects
```

**Description**: Create a new project (org admin or superadmin)  
**Authentication**: Required  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "name": "Project Name"
}
```

**Response** (201 Created):
```json
{
  "id": "proj_456",
  "organization_id": "org_123",
  "name": "Project Name",
  "created_at": "2025-05-29T12:00:00Z",
  "updated_at": "2025-05-29T12:00:00Z"
}
```

#### Get Project
```http
GET /v1/projects/{id}
```

**Description**: Get project details by ID  
**Authentication**: Required

**Response** (200 OK):
```json
{
  "id": "proj_123",
  "organization_id": "org_123",
  "name": "AI Chatbot Project",
  "created_at": "2025-05-29T10:00:00Z",
  "updated_at": "2025-05-29T10:00:00Z"
}
```

#### Update Project
```http
PUT /v1/projects/{id}
```

**Description**: Update project details (org admin or superadmin)  
**Authentication**: Required  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "name": "Updated Project Name"
}
```

**Response** (200 OK):
```json
{
  "id": "proj_123",
  "organization_id": "org_123",
  "name": "Updated Project Name",
  "created_at": "2025-05-29T10:00:00Z",
  "updated_at": "2025-05-29T12:30:00Z"
}
```

#### Delete Project
```http
DELETE /v1/projects/{id}
```

**Description**: Delete project (org admin or superadmin)  
**Authentication**: Required

**Response**: `204 No Content`

---

### API Key Management

#### List API Keys
```http
GET /v1/api-keys
```

**Description**: List API keys (filtered by user access)  
**Authentication**: Required

**Response** (200 OK):
```json
[
  {
    "id": "key_123",
    "organization_id": "org_123",
    "project_id": "proj_123",
    "provider": "openai",
    "api_key": "sk-***masked***",
    "created_at": "2025-05-29T10:00:00Z",
    "updated_at": "2025-05-29T10:00:00Z"
  }
]
```

#### Create API Key
```http
POST /v1/api-keys
```

**Description**: Create/store API key for a provider (org admin or superadmin)  
**Authentication**: Required  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "provider": "openai|groq|google|anthropic|mistral",
  "api_key": "actual_api_key_value"
}
```

**Response** (201 Created):
```json
{
  "id": "key_456",
  "organization_id": "org_123",
  "project_id": "proj_123",
  "provider": "openai",
  "api_key": "sk-***masked***",
  "created_at": "2025-05-29T12:00:00Z",
  "updated_at": "2025-05-29T12:00:00Z"
}
```

#### Revoke API Key
```http
DELETE /v1/api-keys/{id}
```

**Description**: Delete/revoke API key (org admin or superadmin)  
**Authentication**: Required

**Response**: `204 No Content`

---

## Prometheus Monitoring

### Available Metrics

#### LLM Request Latency
```
llm_request_latency_seconds{provider="openai", model="gpt-4"}
```
**Type**: Histogram  
**Description**: Latency of LLM requests in seconds  
**Labels**: `provider`, `model`

#### LLM Token Usage
```
llm_tokens_total{provider="openai", model="gpt-4", type="prompt"}
llm_tokens_total{provider="openai", model="gpt-4", type="completion"}
llm_tokens_total{provider="openai", model="gpt-4", type="total"}
```
**Type**: Counter  
**Description**: Total number of tokens used  
**Labels**: `provider`, `model`, `type`

#### LLM Request Errors
```
llm_request_errors_total{provider="openai", model="gpt-4"}
```
**Type**: Counter  
**Description**: Total number of LLM request errors  
**Labels**: `provider`, `model`

#### API Request Latency
```
api_request_latency_seconds{endpoint="/v1/llm/query", method="POST"}
```
**Type**: Histogram  
**Description**: Latency of API requests in seconds  
**Labels**: `endpoint`, `method`

#### API Request Errors
```
api_request_errors_total{endpoint="/v1/llm/query", method="POST", status="500"}
```
**Type**: Counter  
**Description**: Total number of API request errors  
**Labels**: `endpoint`, `method`, `status`

### Prometheus Queries for Frontend Dashboards

#### Real-time Request Rate
```promql
# Requests per second by provider
rate(llm_tokens_total[5m])

# Success rate percentage
(
  rate(api_request_latency_seconds_count[5m]) - 
  rate(api_request_errors_total[5m])
) / rate(api_request_latency_seconds_count[5m]) * 100
```

#### Performance Metrics
```promql
# Average latency by provider
rate(llm_request_latency_seconds_sum[5m]) / rate(llm_request_latency_seconds_count[5m])

# 95th percentile latency
histogram_quantile(0.95, rate(llm_request_latency_seconds_bucket[5m]))

# Error rate by provider
rate(llm_request_errors_total[5m])
```

#### Usage Analytics
```promql
# Total tokens consumed per hour
increase(llm_tokens_total[1h])

# Cost estimation (tokens * cost_per_1k / 1000)
increase(llm_tokens_total{type="total"}[1h]) * 0.03 / 1000

# Most used models
topk(5, rate(llm_tokens_total[1h]))
```

#### Organization-level Metrics
```promql
# Usage by organization (requires custom labels)
sum by (organization_id) (rate(llm_tokens_total[5m]))

# Project performance comparison
avg by (project_id) (rate(llm_request_latency_seconds_sum[5m]) / rate(llm_request_latency_seconds_count[5m]))
```

### Prometheus Configuration

The service is configured to be scraped by Prometheus:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gateway'
    static_configs:
      - targets: ['app:8080']
    metrics_path: '/metrics'
```

**Prometheus UI**: http://localhost:9090

---

## Error Handling

### Standard Error Response Format
```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional error details"
  }
}
```

### Common HTTP Status Codes

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `204 No Content`: Successful deletion
- `400 Bad Request`: Invalid request format or parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: LLM provider unavailable

---

## Rate Limiting

The service implements rate limiting based on:
- **Per User**: 100 requests per minute
- **Per Organization**: 1000 requests per minute  
- **Per Project**: 500 requests per minute

Rate limit headers included in responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1653900000
```

---

## Supported LLM Providers

### OpenAI
- **Models**: `gpt-4`, `gpt-3.5-turbo`
- **Features**: Chat completions, streaming
- **API Key Format**: `sk-...`

### Groq
- **Models**: `mixtral-8x7b-32768`, `llama2-70b-4096`
- **Features**: Ultra-fast inference
- **API Key Format**: `gsk_...`

### Google AI Studio
- **Models**: `gemini-pro`, `gemini-pro-vision`
- **Features**: Multimodal capabilities
- **API Key Format**: `AIza...`

### Anthropic
- **Models**: `claude-3-opus`, `claude-3-sonnet`
- **Features**: Constitutional AI
- **API Key Format**: `sk-ant-...`

### Mistral
- **Models**: `mistral-large`, `mistral-medium`
- **Features**: European AI provider
- **API Key Format**: Custom format

---

## Intelligence Layer (Optional)

The service includes an optional intelligence layer for response evaluation:

### Configuration
```env
INTELLIGENCE_ENABLED=true
INTELLIGENCE_SAMPLING_RATE=0.1
INTELLIGENCE_KEYWORDS=safety,relevance,completeness
INTELLIGENCE_MIN_SCORE=0.7
INTELLIGENCE_MAX_LATENCY=5s
```

### Evaluation Response
```json
{
  "score": 0.85,
  "reason": "Response is accurate and well-structured",
  "suggestions": ["Consider adding more specific examples"],
  "metadata": {
    "model": "gpt-4",
    "latency": "250ms",
    "token_usage": "35",
    "finish_reason": "stop"
  },
  "timestamp": "2025-05-29T12:00:00Z"
}
```

---

## Frontend Integration Examples

### React/TypeScript Integration

```typescript
// API Client
class GatewayAPIClient {
  private baseURL = 'http://localhost:8080';
  private token: string;

  constructor(clerkToken: string) {
    this.token = clerkToken;
  }

  private async request(endpoint: string, options: RequestInit = {}) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    return response.json();
  }

  // LLM Query
  async queryLLM(request: LLMQueryRequest) {
    return this.request('/v1/llm/query', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // Organization Management
  async getOrganizations() {
    return this.request('/v1/orgs');
  }

  async createOrganization(name: string) {
    return this.request('/v1/orgs', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  // Project Management
  async getProjects() {
    return this.request('/v1/projects');
  }

  // API Key Management
  async getAPIKeys() {
    return this.request('/v1/api-keys');
  }
}
```

### Dashboard Metrics Integration

```typescript
// Prometheus Client for Dashboards
class MetricsClient {
  private prometheusURL = 'http://localhost:9090';

  async queryMetrics(query: string, timeRange = '5m') {
    const response = await fetch(
      `${this.prometheusURL}/api/v1/query?query=${encodeURIComponent(query)}`
    );
    return response.json();
  }

  // Get request rate
  async getRequestRate() {
    return this.queryMetrics('rate(llm_tokens_total[5m])');
  }

  // Get average latency
  async getAverageLatency() {
    return this.queryMetrics(
      'rate(llm_request_latency_seconds_sum[5m]) / rate(llm_request_latency_seconds_count[5m])'
    );
  }

  // Get error rate
  async getErrorRate() {
    return this.queryMetrics('rate(llm_request_errors_total[5m])');
  }
}
```

---

## Security Considerations

### Authentication
- Uses Clerk.dev for robust authentication
- JWT tokens contain organization and project context
- Tokens should be refreshed regularly

### Authorization
- Role-based access control at organization and project levels
- API keys are masked in responses
- Rate limiting prevents abuse

### Data Privacy
- User prompts and responses are not logged by default
- API keys are encrypted at rest
- Audit trails for administrative actions

### Network Security
- HTTPS required in production
- CORS configuration for web frontends
- API key rotation recommended

---

This documentation provides a complete reference for integrating with the LLM Gateway service. The frontend team can use this to build comprehensive dashboards with user management, real-time monitoring, and LLM interaction capabilities.

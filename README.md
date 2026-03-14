# Call Notes AI Service

AI-powered structured note-taking for live medical support calls. Automatically extracts medical entities from real-time call transcripts and populates Salesforce fields — reducing agent post-call work from 3 minutes to near zero.

## Architecture

```
Talkdesk ──→ Audio Stream ──→ Deepgram STT ──→ Transcript Pipeline
                                                        │
                                    ┌───────────────────┤
                                    ▼                   ▼
                              Rule Engine        Medical NER
                              (regex, Hindi      (spaCy, ICD-10)
                               dictionary)             │
                                    │                   │
                                    └────────┬──────────┘
                                             ▼
                                    LLM Reasoning (conditional)
                                    Claude Haiku via Bedrock
                                             │
                                             ▼
                                    Field Mapper ──→ Redis (session state)
                                             │
                                             ▼
                                    WebSocket ──→ Agent UI (Salesforce Canvas)
                                             │
                                             ▼
                                    Salesforce Upsert (on submit)
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24+ |
| HTTP Router | go-chi/chi v5 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Message Queue | Apache Kafka |
| STT | Deepgram Nova-2 |
| LLM | Claude 3.5 Haiku (AWS Bedrock) |
| CRM | Salesforce REST API |
| Metrics | Prometheus |
| Logging | Zap (structured JSON) |
| Tracing | OpenTelemetry + Jaeger |

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.24+ | [golang.org/dl](https://golang.org/dl/) |
| Docker | 20+ | [docs.docker.com](https://docs.docker.com/get-docker/) |
| Docker Compose | v2+ | Included with Docker Desktop |

All other tools (`migrate`, `mockgen`, `golangci-lint`, `goimports`) are installed automatically by `make setup`.

## Quick Start (One Command)

```bash
cd call-notes-ai-service
make setup
```

This runs 5 steps automatically:
1. Installs dev tools (migrate, mockgen, golangci-lint, goimports)
2. Downloads Go dependencies
3. Starts infrastructure (Postgres, Redis, Kafka via Docker)
4. Runs database migrations
5. Builds the binary

Then start the service:

```bash
make run
```

The service starts on two ports:
- **:8080** — Main API (session management, field updates, analytics)
- **:8081** — Ops (health checks, Prometheus metrics)

## Manual Step-by-Step Setup

If you prefer to run each step individually:

```bash
# 1. Install dev tools
make deps-install

# 2. Download Go dependencies
make deps

# 3. Start infrastructure (Postgres, Redis, Kafka)
make docker-up

# 4. Wait for Postgres to be ready, then run migrations
make migrate-up

# 5. Build the binary
make build

# 6. Run the service
make run
```

## Verify It Works

### Quick Smoke Test

```bash
curl http://localhost:8081/health/live
# → {"status":"SERVING"}
```

### Run All API Tests (Automated)

A script tests all 20 endpoints automatically — creates a session, updates fields, submits, tests error handling, hits all futuristic APIs, and purges:

```bash
./scripts/test-apis.sh
```

Expected output:

```
═══════════════════════════════════════════════════
  Call Notes AI Service — API Test Suite
═══════════════════════════════════════════════════

[1/6] Health & Ops
  ✓ GET /health/live (HTTP 200)
  ✓ GET /health/ready (HTTP 200)
  ✓ GET /metrics (HTTP 200)

[2/6] Session — Create
  ✓ POST /v1/sessions (HTTP 201)

[3/6] Session — CRUD
  ✓ GET /v1/sessions/{id} (HTTP 200)
  ✓ PATCH /v1/sessions/{id}/fields (HTTP 200)
  ✓ POST /v1/sessions/{id}/submit (HTTP 200)

[4/6] Session — Error Handling
  ✓ GET /v1/sessions/not-a-uuid → 400
  ✓ GET /v1/sessions/{non-existent} → 404
  ✓ POST /v1/sessions (empty body) → 400

[5/6] Futuristic APIs
  ✓ GET /v1/patients/{phone}/history (HTTP 200)
  ✓ GET /v1/sessions/{id}/triage → 404
  ✓ GET /v1/sessions/{id}/followups (HTTP 200)
  ✓ GET /v1/analytics/overview (HTTP 200)
  ...

[6/6] Session — Purge
  ✓ DELETE /v1/sessions/{id}/purge (HTTP 200)
  ✓ GET /v1/sessions/{id} after purge → 404

═══════════════════════════════════════════════════
  All 20 tests passed
═══════════════════════════════════════════════════
```

### Manual cURL Examples

<details>
<summary>Click to expand all cURL commands</summary>

#### Health & Ops

```bash
# Liveness
curl http://localhost:8081/health/live

# Readiness (checks DB)
curl http://localhost:8081/health/ready

# Prometheus metrics
curl http://localhost:8081/metrics
```

#### Session Lifecycle

```bash
# Create a session
curl -X POST http://localhost:8080/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "talkdesk_call_id": "TK-001",
    "agent_id": "agent-ramesh",
    "patient_phone": "+919876543210"
  }'
# → {"session_id":"<uuid>","status":"CREATED"}

# Get session state
curl http://localhost:8080/v1/sessions/<session_id>

# Update fields (agent overrides)
curl -X PATCH http://localhost:8080/v1/sessions/<session_id>/fields \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"field_name": "patient_name", "value": "Rajesh Kumar"},
      {"field_name": "primary_symptom", "value": "knee pain"},
      {"field_name": "body_part", "value": "right knee"},
      {"field_name": "duration", "value": "2 weeks"},
      {"field_name": "severity", "value": "7/10"}
    ]
  }'

# Submit session (with optional last-minute overrides)
curl -X POST http://localhost:8080/v1/sessions/<session_id>/submit \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"field_name": "medication", "value": "Paracetamol 500mg"}
    ]
  }'

# Purge session (DPDP right-to-erasure)
curl -X DELETE http://localhost:8080/v1/sessions/<session_id>/purge
```

#### Predictive Pre-population

```bash
# Patient history — returns predicted fields for returning patients
curl http://localhost:8080/v1/patients/+919876543210/history
```

#### Triage

```bash
# Get triage assessment for a session
curl http://localhost:8080/v1/sessions/<session_id>/triage
```

#### Follow-ups

```bash
# List detected follow-ups
curl http://localhost:8080/v1/sessions/<session_id>/followups

# Confirm or dismiss a follow-up
curl -X POST http://localhost:8080/v1/sessions/<session_id>/followups/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "followup_id": "<followup_uuid>",
    "confirmed": true,
    "agent_id": "agent-ramesh"
  }'
```

#### Analytics

```bash
# Dashboard overview
curl "http://localhost:8080/v1/analytics/overview?from=2026-03-01&to=2026-03-14"

# Top conditions
curl "http://localhost:8080/v1/analytics/conditions?from=2026-03-01&to=2026-03-14&limit=10"

# Agent performance
curl "http://localhost:8080/v1/analytics/agents/agent-ramesh/performance?from=2026-03-01&to=2026-03-14"

# Sentiment trends
curl "http://localhost:8080/v1/analytics/sentiment?from=2026-03-01&to=2026-03-14&granularity=daily"
```

</details>

## API Reference

### Core APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe (checks DB) |
| GET | `/metrics` | Prometheus metrics |
| POST | `/v1/sessions` | Create a new call session |
| GET | `/v1/sessions/{id}` | Get session state with extracted fields |
| PATCH | `/v1/sessions/{id}/fields` | Agent field overrides |
| POST | `/v1/sessions/{id}/submit` | Submit session to Salesforce |
| DELETE | `/v1/sessions/{id}/purge` | Purge session data (DPDP compliance) |

### Futuristic Feature APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/patients/{phone}/history` | Predicted pre-fill fields for returning patients |
| GET | `/v1/sessions/{id}/triage` | Triage/urgency assessment for session |
| GET | `/v1/sessions/{id}/followups` | Detected follow-up actions for session |
| POST | `/v1/sessions/{id}/followups/confirm` | Confirm or dismiss a detected follow-up |
| GET | `/v1/analytics/overview` | Dashboard overview (calls, accuracy, triage) |
| GET | `/v1/analytics/conditions` | Top conditions with frequency |
| GET | `/v1/analytics/agents/{id}/performance` | Per-agent accuracy and override metrics |
| GET | `/v1/analytics/sentiment` | Sentiment distribution over time |

## Environment Variables

All config lives in `config/default.toml`. Override via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `dev` | Environment (dev/test/prod) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `callnotes` | PostgreSQL database name |
| `DEEPGRAM_API_KEY` | — | Deepgram STT API key |
| `SF_INSTANCE_URL` | — | Salesforce instance URL |
| `SF_CLIENT_ID` | — | Salesforce OAuth client ID |
| `SF_CLIENT_SECRET` | — | Salesforce OAuth client secret |

## Project Structure

```
call-notes-ai-service/
├── cmd/api/main.go                  # Entry point, graceful shutdown
├── config/default.toml              # Default configuration
├── internal/
│   ├── boot/boot.go                 # Application bootstrap, DI wiring
│   ├── config/config.go             # Typed config structs
│   ├── constants/                   # Split constant files (session, http, log, metrics, errors)
│   ├── database/migrations/         # PostgreSQL schema migrations
│   ├── interceptors/                # Full HTTP middleware chain (13 middlewares)
│   │   ├── init.go                  # Chain builder, GetChiMiddlewareWithFullConfig
│   │   ├── http.go                  # Recovery, logging, metrics, security, timeout
│   │   ├── tracing.go              # OpenTelemetry tracing middleware
│   │   └── ratelimit.go            # Token bucket rate limiting
│   ├── logger/                      # Structured zap logging with context
│   ├── metrics/                     # Prometheus metrics definitions
│   ├── tracing/                     # OpenTelemetry tracer setup
│   ├── utils/                       # Shared helpers (JSON, validation)
│   ├── websocket/                   # WebSocket hub for agent UI
│   ├── modules/
│   │   ├── health/                  # Liveness + readiness probes
│   │   ├── session/                 # Call session lifecycle, CRUD
│   │   ├── transcription/           # Transcript chunk processing
│   │   ├── extraction/              # 3-layer NLP pipeline
│   │   ├── fieldmapper/             # Entity → Salesforce field mapping
│   │   ├── reasoning/               # LLM conflict resolution
│   │   ├── salesforce/              # Salesforce upsert orchestration
│   │   ├── prediction/              # Predictive pre-population
│   │   ├── sentiment/               # Voice emotion/sentiment detection
│   │   ├── triage/                  # Predictive triage scoring
│   │   ├── followup/                # Auto follow-up scheduling
│   │   └── analytics/               # Post-call analytics dashboards
│   └── services/
│       ├── deepgram/                # Deepgram WebSocket STT client
│       ├── llm/                     # AWS Bedrock Claude client
│       └── sfdc/                    # Salesforce REST API client
├── pkg/
│   ├── apperror/                    # Typed application errors (codes, responses)
│   ├── config/                      # Config loader with env var template expansion
│   └── database/                    # PostgreSQL pool management (IPool interface)
├── deployment/dev/                  # Docker Compose (Postgres, Redis, Kafka)
├── Dockerfile                       # Multi-stage production build
├── Makefile                         # Build, test, docker, migrate, trace targets
└── go.mod
```

Each module follows the pattern: `init.go` → `core.go` → `server.go` → `repository.go` → `entities/`

## Database

```bash
make migrate-up         # Run all pending migrations
make migrate-down-one   # Rollback one migration
make migrate-create     # Create a new migration file
make migrate-version    # Check current version
make migrate-force      # Force a specific version (for stuck migrations)
```

### Schema

| Table | Purpose |
|-------|---------|
| `call_sessions` | Call session lifecycle (status, agent, timestamps) |
| `extracted_fields` | Versioned field extractions with confidence + source |
| `agent_overrides` | Audit trail of agent corrections |
| `audit_logs` | General audit events (JSONB details) |
| `patient_history_cache` | Patient field history for predictive pre-population |
| `sentiment_logs` | Per-segment emotion detection logs |
| `triage_assessments` | Session urgency scores with symptoms and modifiers |
| `follow_ups` | Detected and confirmed follow-up actions |

## Tracing

For local distributed tracing with Jaeger:

```bash
make trace-up    # Start Jaeger (UI: http://localhost:16686)
make trace-down  # Stop Jaeger
```

Then enable tracing in `config/default.toml`:

```toml
[tracing]
enabled = true
endpoint = "localhost:4317"
sample_rate = 1.0
insecure = true
```

## Makefile Targets

```bash
make help    # Show all available commands with descriptions
```

| Category | Command | Description |
|----------|---------|-------------|
| **Setup** | `make setup` | First-time setup (tools + deps + docker + migrate + build) |
| **Build** | `make build` | Build binary to `bin/` |
| | `make run` | Run the service |
| | `make run-dev` | Run with hot reload (requires `air`) |
| **Test** | `make test` | Run all tests (verbose, race detector) |
| | `make test-short` | Quick test run |
| | `make test-coverage` | Generate coverage report |
| **Quality** | `make lint` | Run golangci-lint |
| | `make vet` | Run go vet |
| | `make check` | Run vet + lint + test |
| | `make fmt` | Format code |
| **Docker** | `make docker-up` | Start Postgres + Redis + Kafka |
| | `make docker-down` | Stop containers |
| | `make docker-clean` | Stop + remove volumes |
| | `make docker-build` | Build Docker image |
| | `make docker-run` | Run Docker container |
| **Database** | `make migrate-up` | Run migrations |
| | `make migrate-down` | Rollback all |
| | `make migrate-force` | Force migration version |
| **Tracing** | `make trace-up` | Start Jaeger |
| | `make trace-down` | Stop Jaeger |
| **Mocks** | `make mock` | Generate test mocks |

## Extraction Pipeline

The 3-layer hybrid NLP pipeline minimizes LLM cost while maximizing accuracy:

**Layer 1 — Rule Engine** (cost: $0, latency: <10ms)
- Phone number regex (Indian mobile: `[6-9]\d{9}`)
- Age extraction (`45 saal`, `32 years old`)
- Duration parsing (`2 hafte` → `2 weeks`)
- Severity scale (`7/10`, `bahut tez`)
- Hindi medical dictionary (30+ symptom mappings)
- Negation detection (`nahi`, `no`, `bina`)

**Layer 2 — Medical NER** (cost: $0 self-hosted, latency: <50ms)
- spaCy + scispaCy for English medical entities
- Custom Hindi medical terms dictionary
- ICD-10 / SNOMED code mapping

**Layer 3 — LLM Reasoning** (cost: ~$0.001/call, latency: <500ms)
- Only invoked when needed: corrections, contradictions, ambiguity
- Transcript grounding: every output must trace to transcript text
- Claude 3.5 Haiku via AWS Bedrock (GPT-4o-mini fallback)

## Troubleshooting

### Docker containers won't start
```bash
make docker-clean    # Remove old volumes
make docker-up       # Start fresh
```

### Migration stuck / dirty state
```bash
make migrate-force   # Enter the version number to force
make migrate-up      # Re-run migrations
```

### Port already in use
```bash
lsof -i :8080       # Find the process
kill -9 <PID>        # Kill it
```

### Redis connection failed
The service continues without Redis — session caching is disabled but all other features work. Check Redis is running:
```bash
docker ps | grep redis
```

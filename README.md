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
| Tracing | OpenTelemetry |

## Prerequisites

- **Go** 1.24+
- **Docker** & Docker Compose
- **migrate** CLI (installed automatically via `make deps-install`)

## Quick Start

```bash
# 1. Clone and enter directory
cd call-notes-ai-service

# 2. Install dev tools (migrate, mockgen, golangci-lint)
make deps-install

# 3. Download Go dependencies
make deps

# 4. Start infrastructure (Postgres, Redis, Kafka)
make docker-up

# 5. Run database migrations
make migrate-up

# 6. Start the service
make run
```

The service starts on two ports:
- **:8080** — Main API (session management, field updates)
- **:8081** — Ops (health checks, Prometheus metrics)

## Verify It Works

### Core APIs

```bash
# Health check
curl http://localhost:8081/health/live
# → {"status":"SERVING"}

curl http://localhost:8081/health/ready
# → {"status":"SERVING"}

# Create a call session
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
# → {"session_id":"...","status":"CREATED","fields":{},"transcript_len":0}

# Update fields (agent review / AI extraction)
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
# → {"status":"updated"}

# Submit session (triggers Salesforce upsert)
curl -X POST http://localhost:8080/v1/sessions/<session_id>/submit \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"field_name": "medication", "value": "Paracetamol 500mg"}
    ]
  }'
# → {"session_id":"<uuid>","sf_record_id":"","status":"SUBMITTED"}

# Prometheus metrics
curl http://localhost:8081/metrics
```

### Futuristic Feature APIs

```bash
# Predictive Pre-population — returns patient history for returning patients
curl http://localhost:8080/v1/patients/+919876543210/history
# → {"patient_phone":"+919876543210","total_sessions":1,"predicted_fields":[]}

# Triage Assessment — returns urgency scoring for a session
curl http://localhost:8080/v1/sessions/<session_id>/triage
# → {"error":"triage assessment not found"}  (none created yet — populated during live calls)

# Follow-ups — returns detected follow-up actions for a session
curl http://localhost:8080/v1/sessions/<session_id>/followups
# → {"followups":[]}

# Follow-up Confirm — agent confirms or dismisses a detected follow-up
curl -X POST http://localhost:8080/v1/sessions/<session_id>/followups/confirm \
  -H "Content-Type: application/json" \
  -d '{"followup_id": "<followup_uuid>", "confirmed": true, "agent_id": "agent-ramesh"}'

# Analytics Overview — dashboard summary for a date range
curl "http://localhost:8080/v1/analytics/overview?from=2026-03-01&to=2026-03-13"
# → {"time_range":{...},"total_calls":3,"avg_call_duration_min":0,...}

# Analytics — Top conditions by frequency
curl "http://localhost:8080/v1/analytics/conditions?from=2026-03-01&to=2026-03-13&limit=10"
# → {"time_range":{...},"conditions":[{"condition":"knee pain","count":2,"pct":0}],"total":2}

# Analytics — Agent performance metrics
curl "http://localhost:8080/v1/analytics/agents/agent-ramesh/performance?from=2026-03-01&to=2026-03-13"
# → {"agent_id":"agent-ramesh","total_calls":2,...}

# Analytics — Sentiment trends over time
curl "http://localhost:8080/v1/analytics/sentiment?from=2026-03-01&to=2026-03-13&granularity=daily"
# → {"time_range":{...},"granularity":"daily","data_points":[]}
```

## API Endpoints

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

### Futuristic Feature APIs
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/patients/{phone}/history` | Get predicted pre-fill fields for returning patient |
| GET | `/v1/sessions/{id}/triage` | Get triage/urgency assessment for session |
| GET | `/v1/sessions/{id}/followups` | List detected follow-ups for session |
| POST | `/v1/sessions/{id}/followups/confirm` | Confirm or dismiss a detected follow-up |
| GET | `/v1/analytics/overview` | Dashboard overview (calls, accuracy, triage, sentiment) |
| GET | `/v1/analytics/conditions` | Top conditions with frequency trends |
| GET | `/v1/analytics/agents/{id}/performance` | Per-agent accuracy and override metrics |
| GET | `/v1/analytics/sentiment` | Sentiment distribution over time |

## Project Structure

```
call-notes-ai-service/
├── cmd/api/main.go                  # Entry point, graceful shutdown
├── config/default.toml              # Default configuration
├── internal/
│   ├── boot/boot.go                 # Application bootstrap, DI wiring
│   ├── config/config.go             # Typed config structs
│   ├── constants/                   # String constants, context keys
│   ├── database/migrations/         # PostgreSQL schema migrations
│   ├── interceptors/                # HTTP middleware chain
│   ├── logger/                      # Structured zap logging
│   ├── metrics/                     # Prometheus metrics definitions
│   ├── websocket/                   # WebSocket hub for agent UI
│   ├── modules/
│   │   ├── health/                  # Liveness + readiness probes
│   │   ├── session/                 # Call session lifecycle, CRUD
│   │   ├── transcription/           # Transcript chunk processing
│   │   ├── extraction/              # 3-layer NLP pipeline
│   │   │   ├── rule_engine.go       #   L1: Regex + Hindi dictionary
│   │   │   ├── llm_reasoner.go      #   L3: Conditional LLM reasoning
│   │   │   └── core.go              #   Pipeline orchestrator
│   │   ├── prediction/              # Predictive pre-population (history)
│   │   ├── sentiment/               # Voice emotion/sentiment detection
│   │   ├── triage/                  # Predictive triage scoring
│   │   ├── followup/                # Auto follow-up scheduling
│   │   ├── analytics/               # Post-call analytics dashboards
│   │   ├── fieldmapper/             # Entity → Salesforce field mapping
│   │   ├── reasoning/               # LLM conflict resolution
│   │   └── salesforce/              # Salesforce upsert orchestration
│   └── services/
│       ├── deepgram/                # Deepgram WebSocket STT client
│       ├── llm/                     # AWS Bedrock Claude client
│       └── sfdc/                    # Salesforce REST API client
├── pkg/
│   ├── apperror/                    # Typed application errors
│   └── database/                    # PostgreSQL pool management
├── deployment/dev/                  # Docker Compose (Postgres, Redis, Kafka)
├── Dockerfile                       # Multi-stage production build
├── Makefile                         # Build, test, docker, migrate targets
└── go.mod
```

Each module follows the pattern: `init.go` → `core.go` → `server.go` → `repository.go` → `entities/`

## Configuration

All config lives in `config/default.toml`. Override via environment variables:

```bash
# Database
export DB_HOST=myhost
export DB_PASSWORD=secret

# Deepgram
export DEEPGRAM_API_KEY=your-key

# Salesforce
export SF_INSTANCE_URL=https://myorg.salesforce.com
export SF_CLIENT_ID=...
export SF_CLIENT_SECRET=...
```

## Database

```bash
# Run migrations
make migrate-up

# Rollback one migration
make migrate-down-one

# Create a new migration
make migrate-create
# Enter name: add_transcript_table

# Check current version
make migrate-version
```

### Schema

| Table | Purpose |
|-------|---------|
| `call_sessions` | Call session lifecycle (status, agent, timestamps) |
| `extracted_fields` | Versioned field extractions with confidence + source |
| `agent_overrides` | Audit trail of agent corrections (AI value vs agent value) |
| `audit_logs` | General audit events (JSONB details) |
| `patient_history_cache` | Patient field history for predictive pre-population |
| `sentiment_logs` | Per-segment emotion detection logs |
| `triage_assessments` | Session urgency scores with symptoms and modifiers |
| `follow_ups` | Detected and confirmed follow-up actions |

## Docker

```bash
make docker-up       # Start Postgres + Redis + Kafka
make docker-down     # Stop containers
make docker-clean    # Stop + remove volumes
make docker-status   # Check container health
make docker-logs     # Tail all logs
```

## Testing

```bash
make test            # Run all tests (verbose, race detector)
make test-short      # Quick run
make test-coverage   # Generate coverage report (coverage.html)
```

## Makefile Targets

```bash
make help            # Show all available commands
make setup           # First-time setup (tools + deps + docker + migrate)
make build           # Build binary to bin/
make run             # Run the service
make lint            # Run golangci-lint
make fmt             # Format code
make mock            # Generate test mocks
```

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
- **Only invoked when needed**: corrections, contradictions, ambiguity
- Transcript grounding: every output must trace to transcript text
- Claude 3.5 Haiku via AWS Bedrock (GPT-4o-mini fallback)

## Documentation

- **Core Technical Spec**: `docs/call-notes-ai-tech-spec.md`
- **Futuristic Scope Spec**: `docs/call-notes-ai-futuristic-scope.md`

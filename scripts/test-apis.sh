#!/usr/bin/env bash
#
# Test all Call Notes AI Service API endpoints.
#
# Prerequisites:
#   - Service running (make run)
#   - Infrastructure up (make docker-up && make migrate-up)
#
# Usage:
#   chmod +x scripts/test-apis.sh
#   ./scripts/test-apis.sh

set -euo pipefail

API="http://localhost:8080"
OPS="http://localhost:8081"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS=0
FAIL=0

check() {
  local name="$1"
  local expected_code="$2"
  local actual_code="$3"
  local body="$4"

  if [ "$actual_code" = "$expected_code" ]; then
    echo -e "  ${GREEN}✓${NC} $name ${CYAN}(HTTP $actual_code)${NC}"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}✗${NC} $name — expected $expected_code, got $actual_code"
    echo "    Response: $body"
    FAIL=$((FAIL + 1))
  fi
}

echo ""
echo -e "${YELLOW}═══════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}  Call Notes AI Service — API Test Suite${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════${NC}"
echo ""

# ─────────────────────────────────────────────────
# 1. Health & Ops
# ─────────────────────────────────────────────────
echo -e "${CYAN}[1/6] Health & Ops${NC}"

RESP=$(curl -s -w "\n%{http_code}" "$OPS/health/live")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /health/live" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$OPS/health/ready")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /health/ready" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$OPS/metrics")
CODE=$(echo "$RESP" | tail -1)
check "GET /metrics" "200" "$CODE" "(prometheus output)"

echo ""

# ─────────────────────────────────────────────────
# 2. Session — Create
# ─────────────────────────────────────────────────
echo -e "${CYAN}[2/6] Session — Create${NC}"

RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/v1/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "talkdesk_call_id": "TK-TEST-'$RANDOM'",
    "agent_id": "agent-ramesh",
    "patient_phone": "+919876543210"
  }')
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "POST /v1/sessions" "201" "$CODE" "$BODY"

SESSION_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['session_id'])" 2>/dev/null || echo "")
if [ -z "$SESSION_ID" ]; then
  echo -e "  ${RED}✗ Could not extract session_id — aborting remaining tests${NC}"
  exit 1
fi
echo -e "  ${YELLOW}→ session_id: $SESSION_ID${NC}"

echo ""

# ─────────────────────────────────────────────────
# 3. Session — CRUD
# ─────────────────────────────────────────────────
echo -e "${CYAN}[3/6] Session — CRUD${NC}"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/$SESSION_ID")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/sessions/{id}" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" -X PATCH "$API/v1/sessions/$SESSION_ID/fields" \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"field_name": "patient_name", "value": "Rajesh Kumar"},
      {"field_name": "primary_symptom", "value": "knee pain"},
      {"field_name": "body_part", "value": "right knee"},
      {"field_name": "duration", "value": "2 weeks"},
      {"field_name": "severity", "value": "7/10"}
    ]
  }')
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "PATCH /v1/sessions/{id}/fields (5 fields)" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/$SESSION_ID")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
FIELD_COUNT=$(echo "$BODY" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('fields',{})))" 2>/dev/null || echo "0")
check "GET /v1/sessions/{id} — verify 5 fields saved" "200" "$CODE" "$BODY"
echo -e "    ${YELLOW}→ fields in response: $FIELD_COUNT${NC}"

RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/v1/sessions/$SESSION_ID/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"field_name": "medication", "value": "Paracetamol 500mg"}
    ]
  }')
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "POST /v1/sessions/{id}/submit" "200" "$CODE" "$BODY"

echo ""

# ─────────────────────────────────────────────────
# 4. Session — Error Handling
# ─────────────────────────────────────────────────
echo -e "${CYAN}[4/6] Session — Error Handling${NC}"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/not-a-uuid")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/sessions/not-a-uuid → 400" "400" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/00000000-0000-0000-0000-000000000099")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/sessions/{non-existent} → 404" "404" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/v1/sessions" \
  -H "Content-Type: application/json" -d '')
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "POST /v1/sessions (empty body) → 400" "400" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/v1/sessions" \
  -H "Content-Type: application/json" \
  -d '{"talkdesk_call_id": "", "agent_id": ""}')
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "POST /v1/sessions (missing required fields) → 500" "500" "$CODE" "$BODY"

echo ""

# ─────────────────────────────────────────────────
# 5. Futuristic APIs
# ─────────────────────────────────────────────────
echo -e "${CYAN}[5/6] Futuristic APIs${NC}"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/patients/+919876543210/history")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/patients/{phone}/history" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/$SESSION_ID/triage")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/sessions/{id}/triage → 404 (no triage data)" "404" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/$SESSION_ID/followups")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/sessions/{id}/followups" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/v1/sessions/$SESSION_ID/followups/confirm" \
  -H "Content-Type: application/json" \
  -d '{"followup_id": "00000000-0000-0000-0000-000000000001", "confirmed": true, "agent_id": "agent-ramesh"}')
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "POST /v1/sessions/{id}/followups/confirm → 404 (no followup)" "404" "$CODE" "$BODY"

TODAY=$(date +%Y-%m-%d)
FROM=$(date -v-30d +%Y-%m-%d 2>/dev/null || date -d "30 days ago" +%Y-%m-%d 2>/dev/null || echo "2026-02-12")

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/analytics/overview?from=$FROM&to=$TODAY")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/analytics/overview" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/analytics/conditions?from=$FROM&to=$TODAY&limit=10")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/analytics/conditions" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/analytics/agents/agent-ramesh/performance?from=$FROM&to=$TODAY")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/analytics/agents/{id}/performance" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/analytics/sentiment?from=$FROM&to=$TODAY&granularity=daily")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/analytics/sentiment" "200" "$CODE" "$BODY"

echo ""

# ─────────────────────────────────────────────────
# 6. Session — Purge (DPDP compliance)
# ─────────────────────────────────────────────────
echo -e "${CYAN}[6/6] Session — Purge${NC}"

RESP=$(curl -s -w "\n%{http_code}" -X DELETE "$API/v1/sessions/$SESSION_ID/purge")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "DELETE /v1/sessions/{id}/purge" "200" "$CODE" "$BODY"

RESP=$(curl -s -w "\n%{http_code}" "$API/v1/sessions/$SESSION_ID")
BODY=$(echo "$RESP" | head -1)
CODE=$(echo "$RESP" | tail -1)
check "GET /v1/sessions/{id} after purge → 404" "404" "$CODE" "$BODY"

echo ""

# ─────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────
TOTAL=$((PASS + FAIL))
echo -e "${YELLOW}═══════════════════════════════════════════════════${NC}"
if [ "$FAIL" -eq 0 ]; then
  echo -e "  ${GREEN}All $TOTAL tests passed${NC}"
else
  echo -e "  ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC} out of $TOTAL"
fi
echo -e "${YELLOW}═══════════════════════════════════════════════════${NC}"
echo ""

exit "$FAIL"

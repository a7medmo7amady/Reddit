#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# test_api.sh — End-to-end curl tests for the Chat Service API.
#
# Usage:
#   ./test_api.sh              # start infra, run tests, leave running
#   ./test_api.sh --cleanup    # same, but tear down containers after
#   ./test_api.sh --skip-start # assume services are already running
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.test.yml"
PROJECT="chat-test"
BASE_URL="http://localhost:8085/api"
CLEANUP=false
SKIP_START=false

for arg in "$@"; do
  case "$arg" in
    --cleanup)   CLEANUP=true ;;
    --skip-start) SKIP_START=true ;;
  esac
done

# ── Colours ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

PASS=0
FAIL=0
TOTAL=0

# ── Helpers ──────────────────────────────────────────────────────────────────
log()  { echo -e "${CYAN}▸${NC} $*"; }
ok()   { ((PASS+=1)); ((TOTAL+=1)); echo -e "  ${GREEN}✓ PASS${NC}  $1"; }
fail() { ((FAIL+=1)); ((TOTAL+=1)); echo -e "  ${RED}✗ FAIL${NC}  $1 — $2"; }

# assert_status <test_name> <expected_code> <actual_code> [body]
assert_status() {
  local name="$1" want="$2" got="$3"
  if [[ "$got" == "$want" ]]; then
    ok "$name (HTTP $got)"
  else
    fail "$name" "expected $want, got $got${4:+ — $4}"
  fi
}

# curl wrapper that returns "status_code\nbody"
api() {
  local method="$1" path="$2" user="${3:-}" data="${4:-}"
  local -a args=(-s -w '\n%{http_code}' -X "$method")
  args+=(-H "Content-Type: application/json")
  [[ -n "$user" ]] && args+=(-H "X-User-Id: $user")
  [[ -n "$data" ]] && args+=(-d "$data")
  curl "${args[@]}" "${BASE_URL}${path}"
}

# Parse response: sets BODY and CODE
parse() {
  local response="$1"
  CODE="$(echo "$response" | tail -1)"
  BODY="$(echo "$response" | sed '$d')"
}

# Extract JSON field (simple jq wrapper)
json() { echo "$BODY" | jq -r "$1" 2>/dev/null; }

# ── Start infrastructure ─────────────────────────────────────────────────────
if [[ "$SKIP_START" == false ]]; then
  log "Starting test infrastructure (project: ${BOLD}$PROJECT${NC})..."
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up --build -d 2>&1 | tail -3

  log "Waiting for chat-service to be healthy..."
  ATTEMPTS=0
  MAX=40
  until curl -sf http://localhost:8085/health >/dev/null 2>&1; do
    ((ATTEMPTS+=1))
    if (( ATTEMPTS >= MAX )); then
      echo -e "${RED}Chat service did not become healthy after ${MAX}s${NC}"
      docker compose -p "$PROJECT" -f "$COMPOSE_FILE" logs chat-service-test | tail -30
      exit 1
    fi
    sleep 1
  done
  log "Chat service is ${GREEN}healthy${NC}."
fi

echo ""
echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}  Chat Service API Tests${NC}"
echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
echo ""

# ── 0. Health check ──────────────────────────────────────────────────────────
log "Health check"
parse "$(curl -s -w '\n%{http_code}' http://localhost:8085/health)"
assert_status "GET /health" 200 "$CODE"

# ── 1. Missing auth → 401 ───────────────────────────────────────────────────
log "Auth enforcement"
parse "$(api GET /inbox)"
assert_status "GET /inbox without X-User-Id" 401 "$CODE"

# ── 2. Create conversation ───────────────────────────────────────────────────
log "Create conversation"
parse "$(api POST /conversations "user-a" '{"participantIds":["user-b"]}')"
assert_status "POST /conversations" 201 "$CODE"
CONV_ID=$(json '.id')
log "  conversation id = ${YELLOW}$CONV_ID${NC}"

# Bad request (no participants)
parse "$(api POST /conversations "user-a" '{}')"
assert_status "POST /conversations (bad body)" 400 "$CODE"

# ── 3. Send message ─────────────────────────────────────────────────────────
log "Send message"
parse "$(api POST /messages "user-a" "{\"conversationId\":\"$CONV_ID\",\"content\":\"Hello from the test script!\"}")"
assert_status "POST /messages" 201 "$CODE"
MSG_ID=$(json '.id')
log "  message id = ${YELLOW}$MSG_ID${NC}"

# Send another message
parse "$(api POST /messages "user-b" "{\"conversationId\":\"$CONV_ID\",\"content\":\"Hey back!\"}")"
assert_status "POST /messages (user-b)" 201 "$CODE"
MSG_ID_B=$(json '.id')

# Not a participant
parse "$(api POST /messages "outsider" "{\"conversationId\":\"$CONV_ID\",\"content\":\"nope\"}")"
assert_status "POST /messages (not participant)" 400 "$CODE"

# Invalid conversation ID
parse "$(api POST /messages "user-a" '{"conversationId":"bad","content":"x"}')"
assert_status "POST /messages (bad conv id)" 400 "$CODE"

# ── 4. Get messages ─────────────────────────────────────────────────────────
log "Get messages"
parse "$(api GET "/conversations/$CONV_ID/messages" "user-a")"
assert_status "GET /conversations/:id/messages" 200 "$CODE"
MSG_COUNT=$(echo "$BODY" | jq 'length')
if [[ "$MSG_COUNT" -ge 2 ]]; then
  ok "message count ≥ 2 (got $MSG_COUNT)"
else
  fail "message count" "expected ≥ 2, got $MSG_COUNT"
fi

# Not a participant
parse "$(api GET "/conversations/$CONV_ID/messages" "nobody")"
assert_status "GET /conversations/:id/messages (not participant)" 400 "$CODE"

# ── 5. Mark read ─────────────────────────────────────────────────────────────
log "Mark read"
parse "$(api POST "/conversations/$CONV_ID/read" "user-a")"
assert_status "POST /conversations/:id/read" 204 "$CODE"

# Not a participant
parse "$(api POST "/conversations/$CONV_ID/read" "stranger")"
assert_status "POST /conversations/:id/read (not participant)" 400 "$CODE"

# ── 6. Get inbox ─────────────────────────────────────────────────────────────
log "Get inbox"
parse "$(api GET /inbox "user-a")"
assert_status "GET /inbox" 200 "$CODE"
INBOX_LEN=$(echo "$BODY" | jq 'length')
if [[ "$INBOX_LEN" -ge 1 ]]; then
  ok "inbox has ≥ 1 item (got $INBOX_LEN)"
else
  fail "inbox items" "expected ≥ 1, got $INBOX_LEN"
fi

# ── 7. Delete message ───────────────────────────────────────────────────────
log "Delete message"

# Not owner → should fail
parse "$(api DELETE "/messages/$MSG_ID" "user-b")"
assert_status "DELETE /messages/:id (not owner)" 400 "$CODE"

# Owner → should succeed
parse "$(api DELETE "/messages/$MSG_ID" "user-a")"
assert_status "DELETE /messages/:id (owner)" 204 "$CODE"

# ── 8. Report message ───────────────────────────────────────────────────────
log "Report message"
parse "$(api POST "/messages/$MSG_ID_B/report" "user-a" '{"reason":"spam content"}')"
assert_status "POST /messages/:id/report" 204 "$CODE"

# Bad body
parse "$(api POST "/messages/$MSG_ID_B/report" "user-a" '{}')"
assert_status "POST /messages/:id/report (no reason)" 400 "$CODE"

# ── 9. Community room ───────────────────────────────────────────────────────
log "Community room"
parse "$(api GET /communities/test-community-1/room "user-c")"
assert_status "GET /communities/:id/room (create)" 200 "$CODE"
ROOM_ID=$(json '.id')
ROOM_TYPE=$(json '.type')
if [[ "$ROOM_TYPE" == "community" ]]; then
  ok "room type = community"
else
  fail "room type" "expected community, got $ROOM_TYPE"
fi

# Second call → same room
parse "$(api GET /communities/test-community-1/room "user-d")"
assert_status "GET /communities/:id/room (idempotent)" 200 "$CODE"
ROOM_ID_2=$(json '.id')
if [[ "$ROOM_ID" == "$ROOM_ID_2" ]]; then
  ok "same room returned on second call"
else
  fail "room idempotency" "$ROOM_ID vs $ROOM_ID_2"
fi

# Send a message to the community room
parse "$(api POST /messages "user-c" "{\"conversationId\":\"$ROOM_ID\",\"content\":\"community msg\"}")"
assert_status "POST /messages (community room)" 201 "$CODE"

# ── 10. Room messages since ──────────────────────────────────────────────────
log "Room messages since"
SINCE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
sleep 0.1

parse "$(api POST /messages "user-c" "{\"conversationId\":\"$ROOM_ID\",\"content\":\"after-since msg\"}")"
assert_status "POST /messages (after since)" 201 "$CODE"

parse "$(api GET "/chat/$ROOM_ID/messages?since=$SINCE" "user-c")"
assert_status "GET /chat/:room/messages?since=" 200 "$CODE"
SINCE_COUNT=$(echo "$BODY" | jq 'length')
if [[ "$SINCE_COUNT" -ge 1 ]]; then
  ok "messages since filter returned ≥ 1 (got $SINCE_COUNT)"
else
  fail "since filter" "expected ≥ 1, got $SINCE_COUNT"
fi

# ── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
if [[ "$FAIL" -eq 0 ]]; then
  echo -e "  ${GREEN}${BOLD}ALL $TOTAL TESTS PASSED ✓${NC}"
else
  echo -e "  ${RED}${BOLD}$FAIL / $TOTAL TESTS FAILED ✗${NC}"
fi
echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
echo ""

# ── Cleanup ──────────────────────────────────────────────────────────────────
if [[ "$CLEANUP" == true ]]; then
  log "Tearing down test containers..."
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down -v --remove-orphans 2>&1 | tail -3
  log "Done."
fi

exit "$FAIL"

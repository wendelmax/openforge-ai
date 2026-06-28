#!/usr/bin/env bash
# OpenForge Smoke Test — generates smoke-report.json with metrics
set -euo pipefail

GO=${GO:-go}
OPENFORGE=${OPENFORGE:-./openforge}
PORT=${PORT:-9191}
BASE="http://127.0.0.1:${PORT}"
REPORT="smoke-report.json"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

PASS=0
FAIL=0

cleanup() {
  if [ -n "${SERVER_PID:-}" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

measure() {
  local name="$1" url="$2" method="${3:-GET}" data="${4:-}"
  local meta metrics
  meta=$(mktemp)
  metrics=$(mktemp)

  if [ "$method" = "GET" ]; then
    curl -s -w "HTTP_CODE:%{http_code}\nTIME:%{time_total}\nSIZE:%{size_download}\nSPEED:%{speed_download}" \
      -o "$metrics" "$url" > "$meta" 2>&1 || true
  else
    curl -s -w "HTTP_CODE:%{http_code}\nTIME:%{time_total}\nSIZE:%{size_download}\nSPEED:%{speed_download}" \
      -o "$metrics" -X "$method" -H "Content-Type: application/json" -d "$data" "$url" > "$meta" 2>&1 || true
  fi

  HTTP_CODE=$(grep 'HTTP_CODE:' "$meta" | cut -d: -f2 | tr -d ' ')
  TIME_TOTAL=$(grep 'TIME:' "$meta" | cut -d: -f2 | tr -d ' ')
  SIZE=$(grep 'SIZE:' "$meta" | cut -d: -f2 | tr -d ' ')
  SPEED=$(grep 'SPEED:' "$meta" | cut -d: -f2 | tr -d ' ')

  [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ] && PASS=$((PASS+1)) || FAIL=$((FAIL+1))

  jq -nc \
    --arg name "$name" \
    --arg method "$method" \
    --arg http_code "${HTTP_CODE:-0}" \
    --arg time_sec "${TIME_TOTAL:-0}" \
    --arg size_bytes "${SIZE:-0}" \
    --arg speed "${SPEED:-0}" \
    '{
      name: $name,
      method: $method,
      http_code: ($http_code | tonumber),
      time_sec: ($time_sec | tonumber),
      size_bytes: ($size_bytes | tonumber),
      speed: ($speed | tonumber),
      ts: $ts
    }' --arg ts "$TIMESTAMP"

  rm -f "$meta" "$metrics"
}

echo "=== OpenForge Smoke Test ==="
echo "timestamp: $TIMESTAMP"
echo ""

# 1. Build
echo "=== Build ==="
CGO_ENABLED=0 "$GO" build -o "$OPENFORGE" ./cmd/openforge 2>&1
BUILD_SIZE=$(stat -c%s "$OPENFORGE" 2>/dev/null || echo 0)
echo "binary: ${BUILD_SIZE} bytes"

# 2. Start server
echo "=== Start Server ==="
"$OPENFORGE" serve --port "$PORT" &
SERVER_PID=$!
sleep 3

if ! kill -0 "$SERVER_PID" 2>/dev/null; then
  echo "FAIL: server failed to start"
  exit 1
fi

# 3. Run tests
echo "=== Tests ==="
RESULTS=$(
  measure "health" "$BASE/health"
  measure "v1/health" "$BASE/v1/health"
  measure "chat" "$BASE/v1/chat" "POST" '{"model":"phi-3-mini","messages":[{"role":"user","content":"hello"}]}'
  measure "completion" "$BASE/v1/completion" "POST" '{"model":"phi-3-mini","prompt":"hello world"}'
  measure "embeddings" "$BASE/v1/embeddings" "POST" '{"model":"bge-small","input":["hello world"]}'
  measure "rerank" "$BASE/v1/rerank" "POST" '{"model":"bge-small","query":"ai","documents":["doc1","doc2"]}'
  measure "models" "$BASE/v1/models"
  measure "devices" "$BASE/v1/devices"
  measure "benchmark" "$BASE/v1/benchmark" "POST" '{"model":"phi-3-mini","iterations":3}'
  measure "model/load" "$BASE/v1/model/load" "POST" '{"model_id":"phi-3-mini"}'
  measure "model/unload" "$BASE/v1/model/unload" "POST" '{"model_id":"phi-3-mini"}'
)

jq -nc \
  --arg ts "$TIMESTAMP" \
  --argjson results "[$(echo "$RESULTS" | tr '\n' ',' | sed 's/,$//')]" \
  '{
    smoke_test: "OpenForge",
    timestamp: $ts,
    summary: { pass: $pass, fail: $fail, total: ($pass + $fail) },
    results: $results
  }' \
  --arg pass "$PASS" \
  --arg fail "$FAIL" \
  > "$REPORT"

# 4. Summary
echo ""
echo "=== Summary ==="
echo "PASS: $PASS  FAIL: $FAIL  TOTAL: $((PASS+FAIL))"
jq -r '.results[] | "\(.http_code | tostring | "000"[:3]) \(.name): \(.time_sec)s (\(.size_bytes)B)"' "$REPORT" 2>/dev/null || true
echo ""
echo "Report saved: $REPORT"
echo ""

exit $FAIL

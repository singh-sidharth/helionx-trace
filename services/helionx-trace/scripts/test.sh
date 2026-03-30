#!/bin/bash

set -euo pipefail

URL="http://localhost:8080"

check_server() {
  if ! curl --fail -s "$URL/health" >/dev/null; then
    echo "Server not reachable at $URL"
    echo "Start it first, for example: STORE_BACKEND=memory go run ./cmd/server"
    exit 1
  fi
}

REQUEST_ID="${REQUEST_ID:-}"

resolve_request_id() {
  if [[ -n "$REQUEST_ID" ]]; then
    echo "$REQUEST_ID"
    return
  fi

  echo "req-$(date +%s)"
}

add_events() {
  check_server
  REQUEST_ID="$(resolve_request_id)"
  echo "---- Adding events for $REQUEST_ID ----"
  local events_url="$URL/events"

  curl -s -X POST "$events_url" -H "Content-Type: application/json" -d '{"requestId":"'"$REQUEST_ID"'","service":"api","eventType":"request.received","status":"SUCCESS"}'
  sleep 1
  curl -s -X POST "$events_url" -H "Content-Type: application/json" -d '{"requestId":"'"$REQUEST_ID"'","service":"order","eventType":"order.created","status":"SUCCESS"}'
  sleep 2
  curl -s -X POST "$events_url" -H "Content-Type: application/json" -d '{"requestId":"'"$REQUEST_ID"'","service":"payment","eventType":"charge","status":"FAILED","metadata":{"error":"stripe timeout"}}'
  sleep 3
  curl -s -X POST "$events_url" -H "Content-Type: application/json" -d '{"requestId":"'"$REQUEST_ID"'","service":"payment","eventType":"charge","status":"RETRY"}'
  sleep 2
  curl -s -X POST "$events_url" -H "Content-Type: application/json" -d '{"requestId":"'"$REQUEST_ID"'","service":"payment","eventType":"charge","status":"SUCCESS"}'
  echo
}

show_timeline() {
  check_server
  if [[ -z "$REQUEST_ID" ]]; then
    echo "REQUEST_ID is required for timeline"
    echo "Example: REQUEST_ID=req-123 ./scripts/test.sh timeline"
    exit 1
  fi
  echo "---- Timeline for $REQUEST_ID ----"
  curl -s "$URL/debug/$REQUEST_ID" | jq
}

show_summary() {
  check_server
  if [[ -z "$REQUEST_ID" ]]; then
    echo "REQUEST_ID is required for summary"
    echo "Example: REQUEST_ID=req-123 ./scripts/test.sh summary"
    exit 1
  fi
  echo "---- Summary for $REQUEST_ID ----"
  curl -s "$URL/debug/$REQUEST_ID/summary"
  echo
}

usage() {
  echo "Usage: REQUEST_ID=req-123 ./scripts/test.sh [add|timeline|summary|all]"
  echo ""
  echo "Commands:"
  echo "  add       Add test events only"
  echo "  timeline  Fetch JSON timeline only"
  echo "  summary   Fetch human-readable summary only"
  echo "  all       Add events, then fetch summary and timeline"
  echo ""
  echo "Notes:"
  echo "  - add generates a fresh request ID if REQUEST_ID is not provided"
  echo "  - summary/timeline require REQUEST_ID when run separately"
  echo "  - server must be running on http://localhost:8080"
}

command="${1:-all}"

case "$command" in
  add)
    add_events
    ;;
  timeline)
    show_timeline
    ;;
  summary)
    show_summary
    ;;
  all)
    add_events
    show_summary
    show_timeline
    ;;
  *)
    usage
    exit 1
    ;;
esac
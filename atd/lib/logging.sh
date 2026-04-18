#!/usr/bin/env bash

# Centralized ATD Tools Logging Wrapper
# Usage: source "lib/logging.sh"
#        atd_log "tool_name" "Log msg"

# Find root .atd file by going up until found
ATD_CONFIG=".atd"
CURRENT_DIR="$PWD"
while [[ "$CURRENT_DIR" != "/" ]]; do
  if [[ -f "$CURRENT_DIR/.atd" ]]; then
    ATD_CONFIG="$CURRENT_DIR/.atd"
    break
  fi
  CURRENT_DIR="$(dirname "$CURRENT_DIR")"
done

ATD_LOG_PATH=""
if [[ -f "$ATD_CONFIG" ]] && command -v jq >/dev/null 2>&1; then
  ATD_LOG_PATH=$(jq -r '.logging.log_path // empty' "$ATD_CONFIG" 2>/dev/null || true)
  # Resolve relative to project root
  if [[ -n "$ATD_LOG_PATH" && "$ATD_LOG_PATH" != /* ]]; then
    ATD_LOG_PATH="$(dirname "$ATD_CONFIG")/$ATD_LOG_PATH"
  fi
fi

atd_log() {
  local tool="$1"
  local message="$2"

  if [[ -n "$ATD_LOG_PATH" ]]; then
    mkdir -p "$(dirname "$ATD_LOG_PATH")"
    local timestamp=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
    
    local log_entry
    log_entry=$(jq -n -c \
      --arg ts "$timestamp" \
      --arg tl "$tool" \
      --arg msg "$message" \
      '{time: $ts, tool: $tl, msg: $msg}')
    
    echo "$log_entry" >> "$ATD_LOG_PATH"
  fi
}

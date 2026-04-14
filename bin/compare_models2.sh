#!/usr/bin/env bash
set -uo pipefail

# Usage:
#   ./compare-models.sh "your prompt"
#   ./compare-models.sh "your prompt" 300
#
# Requirements:
#   - bash
#   - jq
#   - perl
#
# Notes:
# - Codex uses --output-last-message, so it returns only the final assistant text.
# - OpenCode uses --format json, then jq extracts only assistant-like text fields.
# - OpenCode Go models use the opencode-go/<model-id> format.
# - Byte limit is applied after cleanup/reconstruction.

PROMPT="${1:-}"
LIMIT="${2:-300}"

if [ -z "$PROMPT" ]; then
  echo "Usage: $0 \"prompt\" [byte_limit]" >&2
  exit 1
fi

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
  echo "Error: byte_limit must be an integer" >&2
  exit 1
fi

for cmd in jq perl; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: missing required command: $cmd" >&2
    exit 1
  fi
done

print_divider() {
  printf '%*s\n' 72 '' | tr ' ' '-'
}

strip_ansi() {
  perl -pe 's/\e\[[0-9;?]*[ -\/]*[@-~]//g; s/\r//g'
}

trim_blank_edges() {
  awk '
    { lines[NR]=$0 }
    END {
      start=1
      while (start<=NR && lines[start] ~ /^[[:space:]]*$/) start++
      end=NR
      while (end>=start && lines[end] ~ /^[[:space:]]*$/) end--
      for (i=start; i<=end; i++) print lines[i]
    }'
}

collapse_blank_lines() {
  awk '
    BEGIN { blank=0 }
    {
      if ($0 ~ /^[[:space:]]*$/) {
        if (blank) next
        blank=1
        print ""
      } else {
        blank=0
        print
      }
    }'
}

show_text() {
  local text="$1"
  if [ -n "$text" ]; then
    printf '%s' "$text" | head -c "$LIMIT"
    printf '\n'
  else
    printf '(no output)\n'
  fi
}

cleanup_plain() {
  strip_ansi | collapse_blank_lines | trim_blank_edges
}

extract_opencode_assistant_text() {
  jq -r '
    def to_text:
      if . == null then empty
      elif type == "string" then .
      elif type == "array" then map(
        if type == "string" then .
        elif type == "object" then (.text // .content // empty)
        else empty end
      ) | join("")
      elif type == "object" then (.text // .content // empty)
      else empty end;

    # Keep only assistant-like events/messages.
    select(
      .role == "assistant" or
      .type == "assistant" or
      .event == "assistant" or
      .kind == "assistant" or
      .message.role == "assistant" or
      (.type == "message" and .message.role == "assistant") or
      (.event == "message" and .message.role == "assistant") or
      (.delta.role == "assistant")
    )
    |
    (
      .text //
      .content //
      .delta.text //
      .delta.content //
      .message.content //
      .message.text //
      .message.delta.text //
      .message.delta.content //
      empty
    )
    | to_text
  ' 2>/dev/null | collapse_blank_lines | trim_blank_edges
}

run_plain() {
  local label="$1"
  shift

  print_divider
  printf '[%s]\n' "$label"

  local out status
  out="$("$@" 2>&1 | cleanup_plain)"
  status=$?

  show_text "$out"

  if [ "$status" -ne 0 ]; then
    printf '\n(exit: %s)\n' "$status"
  fi
  printf '\n'
}

run_codex() {
  local label="$1"
  shift

  print_divider
  printf '[%s]\n' "$label"

  local tmp_out tmp_err out err status
  tmp_out="$(mktemp)"
  tmp_err="$(mktemp)"

  "$@" --output-last-message "$tmp_out" >"$tmp_err" 2>&1
  status=$?

  out="$(cat "$tmp_out" 2>/dev/null | cleanup_plain)"
  err="$(cat "$tmp_err" 2>/dev/null | cleanup_plain)"

  rm -f "$tmp_out" "$tmp_err"

  if [ -n "$out" ]; then
    show_text "$out"
  elif [ -n "$err" ]; then
    show_text "$err"
  else
    printf '(no output)\n'
  fi

  if [ "$status" -ne 0 ]; then
    printf '\n(exit: %s)\n' "$status"
  fi
  printf '\n'
}

run_opencode() {
  local label="$1"
  shift

  print_divider
  printf '[%s]\n' "$label"

  local tmp_json tmp_err out err status
  tmp_json="$(mktemp)"
  tmp_err="$(mktemp)"

  "$@" --format json >"$tmp_json" 2>"$tmp_err"
  status=$?

  out="$(extract_opencode_assistant_text <"$tmp_json")"
  err="$(cat "$tmp_err" 2>/dev/null | cleanup_plain)"

  rm -f "$tmp_json" "$tmp_err"

  if [ -n "$out" ]; then
    show_text "$out"
  elif [ -n "$err" ]; then
    show_text "$err"
  else
    printf '(no output)\n'
  fi

  if [ "$status" -ne 0 ]; then
    printf '\n(exit: %s)\n' "$status"
  fi
  printf '\n'
}

# Claude
# run_plain "claude sonnet low" claude -p --model sonnet --effort low "$PROMPT"
# run_plain "claude sonnet medium" claude -p --model sonnet --effort medium "$PROMPT"
# run_plain "claude sonnet high" claude -p --model sonnet --effort high "$PROMPT"
#
# run_plain "claude opus low" claude -p --model opus --effort low "$PROMPT"
# run_plain "claude opus medium" claude -p --model opus --effort medium "$PROMPT"
# run_plain "claude opus high" claude -p --model opus --effort high "$PROMPT"
# run_plain "claude opus max" claude -p --model opus --effort max "$PROMPT"

# Codex GPT-5.4
# run_codex "codex gpt-5.4 low" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="low"' "$PROMPT"
# run_codex "codex gpt-5.4 medium" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="medium"' "$PROMPT"
# run_codex "codex gpt-5.4 high" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="high"' "$PROMPT"
# run_codex "codex gpt-5.4 xhigh" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="xhigh"' "$PROMPT"

# OpenCode Go
run_opencode "opencode-go glm-5" opencode run -m opencode-go/glm-5 "$PROMPT"
run_opencode "opencode-go kimi-k2.5" opencode run -m opencode-go/kimi-k2.5 "$PROMPT"
run_opencode "opencode-go minimax-m2.5" opencode run -m opencode-go/minimax-m2.5 "$PROMPT"
run_opencode "opencode-go minimax-m2.7" opencode run -m opencode-go/minimax-m2.7 "$PROMPT"

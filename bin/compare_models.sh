#!/usr/bin/env bash
set -uo pipefail

# Usage:
#   ./compare-models.sh "your prompt"
#   ./compare-models.sh "your prompt" 300
#
# Notes:
# - Byte limit applies to the final cleaned model text shown for each run.
# - Lines containing "sequential-thinking" are removed before byte truncation.
# - Codex uses --output-last-message so you get only the assistant's final answer.
# - Codex uses --skip-git-repo-check so it works outside a trusted repo.
# - OpenCode Go models use the opencode-go/ prefix.
# - This script is Bash. Fine to run from fish as: bash compare-models.sh "prompt" 300

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

strip_ansi() {
  perl -pe 's/\e\[[0-9;?]*[ -\/]*[@-~]//g; s/\r//g'
}

filter_noise() {
  grep -F -v 'sequential-thinking' || true
}

clean_text() {
  strip_ansi | filter_noise
}

print_divider() {
  printf '%*s\n' 72 '' | tr ' ' '-'
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

run_plain() {
  local label="$1"
  shift

  print_divider
  printf '[%s]\n' "$label"

  local tmp status out
  tmp="$(mktemp)"

  "$@" >"$tmp" 2>&1
  status=$?

  out="$(clean_text <"$tmp")"
  rm -f "$tmp"

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

  local tmp_out tmp_err status out err
  tmp_out="$(mktemp)"
  tmp_err="$(mktemp)"

  "$@" --output-last-message "$tmp_out" >"$tmp_err" 2>&1
  status=$?

  out="$(clean_text <"$tmp_out" 2>/dev/null)"
  err="$(clean_text <"$tmp_err" 2>/dev/null)"

  rm -f "$tmp_out" "$tmp_err"

  if [ -n "$out" ]; then
    show_text "$out"
  elif [ -n "$err" ]; then
    # If Codex fails before producing a final assistant message,
    # show the error text instead.
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
#run_plain "claude sonnet low" claude -p --model sonnet --effort low "$PROMPT"
#run_plain "claude sonnet medium" claude -p --model sonnet --effort medium "$PROMPT"
#run_plain "claude sonnet high" claude -p --model sonnet --effort high "$PROMPT"
# run_plain "claude sonnet max" claude -p --model sonnet --effort max "$PROMPT"

# run_plain "claude opus low" claude -p --model opus --effort low "$PROMPT"
# run_plain "claude opus medium" claude -p --model opus --effort medium "$PROMPT"
# run_plain "claude opus high" claude -p --model opus --effort high "$PROMPT"
# run_plain "claude opus max" claude -p --model opus --effort max "$PROMPT"

# Codex: GPT-5.4 with reasoning effort variants
echo "░▒▓ Codex GPT-5.4 Low"
run_codex "codex gpt-5.4 low" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="low"' "$PROMPT"
#run_codex "codex gpt-5.4 medium" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="medium"' "$PROMPT"
#run_codex "codex gpt-5.4 high" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="high"' "$PROMPT"
#run_codex "codex gpt-5.4 xhigh" codex exec --skip-git-repo-check -m gpt-5.4 -c 'model_reasoning_effort="xhigh"' "$PROMPT"

# OpenCode Go
echo "░▒▓ GLM-5"
run_plain "opencode-go glm-5" opencode run -m opencode-go/glm-5 "$PROMPT"
echo "+sequential-thinking MCP:"
run_plain "opencode-go glm-5" opencode run -m opencode-go/glm-5 "use sequential-thinking MCP to: $PROMPT"

echo "░▒▓ Kimi K2.5"
run_plain "opencode-go kimi-k2.5" opencode run -m opencode-go/kimi-k2.5 "$PROMPT"
echo "+sequential-thinking MCP:"
run_plain "opencode-go kimi-k2.5" opencode run -m opencode-go/kimi-k2.5 "use sequential-thinking MCP to: $PROMPT"

echo "░▒▓ MiniMax M2.7"
run_plain "opencode-go minimax-m2.7" opencode run -m opencode-go/minimax-m2.7 "$PROMPT"
echo "+sequential-thinking MCP:"
run_plain "opencode-go minimax-m2.7" opencode run -m opencode-go/minimax-m2.7 "use sequential-thinking MCP to: $PROMPT"

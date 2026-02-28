#!/bin/bash
set -e

HOOK_INPUT=$(cat)

# Activate the flox environment for access to go, jq, etc.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
eval "$(flox activate -d "$SCRIPT_DIR")"

CWD=$(echo "$HOOK_INPUT" | jq -r '.cwd')
STOP_HOOK_ACTIVE=$(echo "$HOOK_INPUT" | jq -r '.stop_hook_active')

# Prevent infinite loops
if [ "$STOP_HOOK_ACTIVE" = "true" ]; then
  exit 0
fi

cd "$CWD" || exit 1

if go test ./... >&2; then
  exit 0
else
  jq -n '{
    "decision": "block",
    "reason": "Tests failed. Fix the failing tests before completing."
  }'
  exit 0
fi

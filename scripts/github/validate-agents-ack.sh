#!/usr/bin/env bash
set -euo pipefail

pr_body=${1:-}

if [[ -z "$pr_body" ]]; then
  echo "Missing PR description; AGENTS.md acknowledgement is required." >&2
  exit 1
fi

pattern='- \[[xX]\] I have read and will follow AGENTS.md relevant to this change\.'

if ! printf '%s' "$pr_body" | grep -Eq "$pattern"; then
  echo "AGENTS.md acknowledgement checkbox is not checked." >&2
  exit 1
fi

#!/usr/bin/env bash
set -euo pipefail

unformatted=$(gofmt -l .)
if [[ -n $unformatted ]]; then
  printf 'Run gofmt -w on these files:\n%s\n' "$unformatted" >&2
  exit 1
fi

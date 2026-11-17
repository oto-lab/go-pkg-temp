#!/usr/bin/env bash
set -euo pipefail

ref=HEAD
if [[ $RETRY == true ]]; then
  ref="refs/tags/$TAG"
fi
printf 'ref=%s\n' "$(git rev-parse --verify "$ref^{commit}")"

#!/usr/bin/env bash
set -euo pipefail

scripts=$(cd "$(dirname "$0")" && pwd)
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
cd "$work"
git init -q
git config user.name 'Release Rules Test'
git config user.email 'release-rules@example.invalid'

manifest() {
  printf 'module github.com/example/library%s\n\ngo 1.26.0\n' "$1" > go.mod
  git add go.mod
  git commit -qm "Set module path to library$1"
}

expect_version() {
  local version=$1 preid=$2 tag=$3 retry=${4:-false} result
  result=$(VERSION="$version" PREID="$preid" bash "$scripts/next-version.sh")
  if [[ $result != "tag=$tag"$'\n'"retry=$retry" ]]; then
    printf 'VERSION=%s PREID=%s: expected %s (retry=%s), got %s\n' "$version" "$preid" "$tag" "$retry" "$result" >&2
    exit 1
  fi
}

expect_error() {
  local version=$1 preid=$2 message=$3 result
  if result=$(VERSION="$version" PREID="$preid" bash "$scripts/next-version.sh" 2>&1) || [[ $result != *"$message"* ]]; then
    printf 'VERSION=%s PREID=%s: expected an error containing %s, got %s\n' "$version" "$preid" "$message" "$result" >&2
    exit 1
  fi
}

manifest ''
git tag v-not-a-version
expect_version minor '' v0.1.0
git tag v0.0.0
expect_version 0.0.0 '' v0.0.0 true
git tag v1.0.0
expect_version minor rc v1.1.0-rc.0
expect_version prerelease rc v1.0.1-rc.0
expect_error prerelease '' 'a prerelease needs an identifier'
expect_error minor RC 'preid must contain lowercase letters only'
expect_error 1.1.0 beta 'leave preid empty'
git tag v1.1.0-beta.2
expect_version prerelease rc v1.1.0-rc.0
git tag v1.1.0-rc.0
expect_version prerelease '' v1.1.0-rc.1
expect_error prerelease beta 'must be newer'
git tag v1.1.0-rc.1
expect_version minor '' v1.1.0
git tag v1.1.0
expect_version 1.2.0-rc '' v1.2.0-rc.0
expect_version 1.2.0-preview '' v1.2.0-preview.0
git tag v1.2.0-rc.1
expect_version 1.2.0-rc '' v1.2.0-rc.2
expect_version v1.2.0-rc.1 '' v1.2.0-rc.1 true
git tag v1.2.0-rc.9
git tag v1.2.0-rc.10
expect_version v1.2.0-rc '' v1.2.0-rc.11
expect_error 1.2.0-rc.9 '' 'must be newer'
git tag v1.4.2
expect_error major '' 'requires a module path ending in /v2'
manifest /v2
expect_version major '' v2.0.0
expect_error 1.5.0 '' 'requires a module path without a /vN suffix'
expect_version 1.4.2 '' v1.4.2 true
tagged_commit=$(git rev-parse 'v1.4.2^{commit}')
[[ $(TAG=v1.4.2 RETRY=true bash "$scripts/select-ref.sh") == "ref=$tagged_commit" ]]
[[ $(TAG=v2.0.0 RETRY=false bash "$scripts/select-ref.sh") == "ref=$(git rev-parse HEAD)" ]]
git tag v2.0.0
expect_version 2.0.0 '' v2.0.0 true
echo 'Release decisions passed.'

#!/usr/bin/env bash
set -euo pipefail

if [[ $(go list -m) != github.com/oto-lab/go-pkg-temp ]]; then
  echo 'Setup smoke test applies only to unconfigured copies.'
  exit 0
fi

work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
git archive HEAD | tar -x -C "$work"
cd "$work"
git init -q
git remote add origin git@github.com:example-org/renamed-lib.git
printf '\n\nA test library\nExample Author\nMIT\n' | go run scripts/setup.go

[[ $(go list -m) == github.com/example-org/renamed-lib ]]
[[ $(go list -f '{{.Name}}') == renamedlib ]]
[[ -f renamedlib.go && -f renamedlib_test.go && ! -e gopkgtemp.go && ! -e gopkgtemp_test.go ]]
grep -Fq 'https://github.com/example-org/renamed-lib' README.md
grep -Fq 'Example Author' README.md
grep -Fq 'A test library' README.md
grep -Fq 'Place your LICENSE file here.' LICENSE
if grep -Er '<(package-(name|description|repo|git|license|author)|module-path|repository-slug)>' README.md CONTRIBUTING.md .github; then
  echo 'Setup left unresolved placeholders.' >&2
  exit 1
fi
go test ./...

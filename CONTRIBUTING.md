# Contributing

Thanks for your interest in improving <package-name>!

## Getting set up

Install Go 1.26 or newer and [golangci-lint v2](https://golangci-lint.run/docs/welcome/install/). Use a Go patch release that still receives security updates. The race detector requires cgo and a supported C compiler. On Windows use a compiler with mingw-w64 runtime version 8 or later.

Clone the repository:

```sh
git clone <package-git>
```

## Commands

| Command | What it does |
| --- | --- |
| `go build ./...` | Build every package |
| `go test -race ./...` | Run tests and examples with the race detector |
| `go test -race -coverprofile=coverage.out ./...` | Run tests and write a coverage profile |
| `go vet ./...` | Run Go's built-in checks |
| `gofmt -l .` | List files that need formatting; CI expects no output |
| `gofmt -w .` | Format Go files |
| `go mod tidy -diff` | Check that module metadata is tidy |
| `golangci-lint run` | Run lint and import formatting checks |
| `go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...` | Check for known vulnerabilities in reachable code |

The template also checks setup with `bash scripts/check-setup.sh` (in a copy of the current commit) and the release rules with `bash scripts/release/check-next-version.sh` (requires Bash, Git and Node.js with npm).

## Conventions

- Keep the public API at the module root. Put implementation packages under `internal/`.
- Start doc comments for exported identifiers with the identifier's name. Add runnable examples to show how callers use the API.
- Wrap errors with `%w` when callers should be able to inspect the underlying error. Use `errors.Is` and `errors.As` to check wrapped errors.
- Return errors for expected failures. Put `context.Context` first in functions that support cancellation, and avoid global mutable state.
- Test public behavior from the external test package. Use table-driven tests when several inputs share the same assertions, and keep fixtures in `testdata/`.
- Apply lint exceptions on the affected line with `//nolint:<linter> // reason`.

## Pull requests

Keep changes focused and add tests for new behavior. Run the checks in the Commands table before opening a pull request.

## Releasing (maintainers)

Run the `release` workflow on `main`. Set `version` to `patch`, `minor`, `major`, `prerelease`, or an explicit version such as `1.2.0` or `v1.2.0-rc.0`. An optional lowercase `preid` starts or continues a prerelease, for example `minor` with `rc` after `v1.0.0` gives `v1.1.0-rc.0`. Omit `preid` for explicit versions; `1.2.0-rc` picks the next number in that series, starting at zero.

The workflow checks the selected commit, pushes its tag, creates a GitHub Release with generated notes, and asks the public Go module proxy to fetch the version. Install a prerelease explicitly, for example `go get <module-path>@v1.2.0-rc.0`.

- For v2 or later, add the matching `/vN` suffix to the module path and update internal imports before running the release. v0 and v1 use the unsuffixed path.
- Raising the minimum Go version requires at least a minor release. Update `go.mod`, the README badge and requirements, and this guide, then mention the change in the release notes.
- To retry after a tag was pushed, start a new dispatch with the latest tag as an explicit version, including the number for a prerelease. The workflow checks that tag's commit and resumes without moving the tag or replacing an existing GitHub Release. Re-running the original bump can calculate a different version.
- Never move a published tag. If a version is broken, tag a new version with a `retract v1.2.0 // reason` directive in `go.mod`, substituting the broken version and the reason.

## License

By contributing, you agree that your contributions are licensed under the [<package-license> License](./LICENSE).

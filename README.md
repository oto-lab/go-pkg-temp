# <package-name>

<package-description>

[![Go Reference](https://pkg.go.dev/badge/<module-path>.svg)](https://pkg.go.dev/<module-path>)
[![CI](<package-repo>/actions/workflows/ci.yml/badge.svg)](<package-repo>/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/<repository-slug>)](./LICENSE)
[![Go 1.26+](https://img.shields.io/badge/Go-1.26%2B-00ADD8)](#requirements)

## Template setup

Create a repository from this template and clone it. Run `go run scripts/setup.go` in the repository root and answer the prompts. The script uses the clone's GitHub origin as the default module path, renames the sample package, fills in repository metadata, and verifies that the new module builds and passes its tests.

Replace the placeholder [LICENSE](./LICENSE) with your chosen license before publishing. Update the sample and this README to describe your library.

## Install

```sh
go get <module-path>@latest
```

## Usage

See the [runnable examples on pkg.go.dev](https://pkg.go.dev/<module-path>#pkg-examples). Run them locally with `go test -run '^Example' ./...`.

## Requirements

Go 1.26 or newer. CI tests the two most recent Go releases.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). The repository can be cloned with `git clone <package-git>`.

## License

[<package-license>](./LICENSE) · <package-author>

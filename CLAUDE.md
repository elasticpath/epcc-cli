# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

EPCC CLI (`epcc`) is a command-line tool for interacting with the Elastic Path Composable Commerce API. Built with Go and the Cobra CLI framework, it dynamically generates CRUD commands from YAML resource definitions.

## Build & Test Commands

```bash
# Build
make build                    # Output: ./bin/epcc

# Run all unit tests
go test -v -cover ./cmd/ ./external/...

# Run a single test
go test -v -run TestName ./cmd/
go test -v -run TestName ./external/packagename/

# Format check (CI enforces this)
gofmt -s -l .

# Format fix
go fmt "./..."

# Smoke tests (require built binary in PATH and API credentials)
export PATH=./bin/:$PATH
./cmd/get-all-smoke-tests.sh
./external/runbooks/run-all-runbooks.sh
```

## Architecture

### Command Generation Pattern

The CLI dynamically generates commands at startup rather than hardcoding each one:

1. **`main.go`** → `cmd.InitializeCmd()` → `cmd.Execute()`
2. **`cmd/root.go`**: `InitializeCmd()` loads resource YAML definitions, then calls `NewCreateCommand()`, `NewGetCommand()`, `NewUpdateCommand()`, `NewDeleteCommand()`, etc. to build the command tree
3. Each `New*Command()` function (in `cmd/create.go`, `cmd/get.go`, etc.) iterates over all resources and generates a subcommand per resource

### Resource Definitions (`external/resources/`)

Resources are defined in `external/resources/yaml/*.yaml` files, embedded via `//go:embed`. Each YAML file maps a resource type to its API endpoints, field definitions, and autofill capabilities. The `Resource` struct in `external/resources/` is the central type that drives command generation, REST calls, and completion.

#### Resources Definitions Vs. OpenAPI Specs

The Resource Definitions predate OpenAPI specs and have historically been incorrect, a lot of work has been done to make them better, and so they are more authoritative. The resource definitions can express different things more easily or harder than the OpenAPI specs,
for instance while there is one platform, the OpenAPI specs are fragmented, and don't semantically link resources, for instance the 

### Request Flow

```
cmd/{create,get,update,delete}.go  →  external/rest/{create,get,update,delete}.go
    →  external/httpclient/  →  external/authentication/ (token management)
    →  HTTP request to API
```

### Key Packages

- **`external/httpclient/`** - HTTP client with rate limiting, retries (5xx, 429), custom headers, URL rewriting, and request/response logging
- **`external/authentication/`** - Multiple auth flows (Client Credentials, Customer Token, Account Management, OIDC) with token caching
- **`external/runbooks/`** - YAML-based action sequences with Go template rendering, variable systems, and parallel execution
- **`external/aliases/`** - Named references to resource IDs for scripting
- **`external/profiles/`** - Context isolation for multiple environments
- **`external/json/`** - JQ integration for output post-processing

### Design Decision: Loose OpenAPI Dependency

OpenAPI specs are included in the repo but the CLI does not depend on them at runtime. Resource definitions are duplicated in the YAML configs. The tool should build and work without specs; they're primarily used for validation in tests.

The resources definitions are designed to simplify interacting with EPCC via the command line, and as such are more concise.

## Code Style

- Go standard formatting (`gofmt -s`), tabs for Go, 2-space indent for YAML
- Tests use `stretchr/testify` (`require` package)
- No linter beyond `gofmt` is configured

## Configuration

API credentials and CLI behavior are controlled via environment variables prefixed with `EPCC_` (defined in `config/config.go`). Key ones: `EPCC_CLIENT_ID`, `EPCC_CLIENT_SECRET`, `EPCC_API_BASE_URL`. Profile-based context isolation is available via `EPCC_PROFILE`.

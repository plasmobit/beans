# Mirrors the tasks in mise.toml for machines without mise.
# Requires go, node and corepack on PATH; pnpm is run through corepack.
# Node 25+ no longer bundles corepack: install it with `npm install -g corepack`.

set positional-arguments

# pnpm 11+ ignores `onlyBuiltDependencies` in frontend/pnpm-workspace.yaml and refuses to install.
pnpm := "corepack pnpm@10"

# Download pnpm without asking, so recipes never stop at an interactive prompt.
export COREPACK_ENABLE_DOWNLOAD_PROMPT := "0"

ldflags := "-X github.com/hmans/beans/internal/version.Version=" + `git describe --tags --always --dirty 2>/dev/null || echo dev` + " -X github.com/hmans/beans/internal/version.Commit=" + `git rev-parse --short HEAD 2>/dev/null || echo unknown` + " -X github.com/hmans/beans/internal/version.Date=" + `date -u +%Y-%m-%dT%H:%M:%SZ`

default: build

# Install dependencies and generate code
setup: deps codegen

deps: deps-backend deps-frontend

deps-backend:
    go mod download

deps-frontend:
    cd frontend && {{pnpm}} install

codegen:
    go generate ./...
    cd frontend && {{pnpm}} codegen

build-frontend:
    cd frontend && {{pnpm}} build

# Copy the frontend build into the directory embedded by the Go binary
build-embed: build-frontend
    rm -rf internal/web/dist
    mkdir -p internal/web/dist
    cp -r frontend/build/* internal/web/dist/
    touch internal/web/dist/.gitkeep

build: codegen build-embed
    go build -ldflags "{{ldflags}}" -o beans ./cmd/beans
    go build -ldflags "{{ldflags}}" -o beans-serve ./cmd/beans-serve
    go build -ldflags "{{ldflags}}" -o beans-tui ./cmd/beans-tui

install: build
    mkdir -p ~/.local/bin
    cp beans ~/.local/bin/beans
    cp beans-serve ~/.local/bin/beans-serve
    cp beans-tui ~/.local/bin/beans-tui

test: codegen test-e2e
    go test ./...

test-e2e: build-embed
    cd frontend && {{pnpm}} test:e2e

# Build and run the beans CLI, e.g. `just beans list --json`
beans *args:
    go run ./cmd/beans "$@"

beans-serve port="22880":
    go build -o /tmp/beans-serve-dev ./cmd/beans-serve && /tmp/beans-serve-dev --port {{port}}

beans-tui *args:
    go run ./cmd/beans-tui "$@"

dev-frontend port="5173":
    cd frontend && {{pnpm}} dev --port {{port}}

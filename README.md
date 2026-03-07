# DataFlow Web

Web UI and server for managing [DataFlow Operator](https://github.com/dataflow-operator/dataflow): view and edit DataFlow manifests, processor logs, and metrics.

## Structure

- **web/** — Vue 3 + Vite frontend (dashboard, manifests, logs, metrics). Build: `cd web && npm install && npm run build`; output in `web/dist/`, copied into the Docker image as `/app/static`.
- **cmd/server/** — GUI server entrypoint.
- **internal/gui/** — HTTP server, static file serving and API (`/api/dataflows`, `/api/logs`, `/api/status`, `/api/namespaces`).

The image is built **separately** from the operator image. On `docker build`, the frontend (Node) is built first, then the Go server; static assets come from the Vue build.

## Building the image

The `github.com/dataflow-operator/dataflow` dependency is fetched from GitHub (main/master branch), no local `replace`. Build from the **dataflow-web** directory:

```bash
# from dataflow-web directory
cd dataflow-web
docker build -t dataflow-web:local .

# or from repo root
docker build -f dataflow-web/Dockerfile -t dataflow-web:local dataflow-web
```

Image contents: binary `/app/server`, static files in `/app/static`. Server listens on port 8080.

## Taskfile (local build and dev)

The repo includes a [Taskfile](https://taskfile.dev). After installing Task (`go install github.com/go-task/task/v3/cmd/task@latest` or via brew/apt):

```bash
task install   # Go deps + npm in web/
task build     # build frontend + static/ + server binary
task run       # run ./server (after build)
task dev       # dev mode: Go server + Vite dev in parallel
task web:test  # frontend tests
task docker    # build Docker image
task --list    # list tasks
```

## Running locally

**Option 1 — Go only (requires pre-built static):**

```bash
cd dataflow-web
# Build frontend and copy to static (or use the image)
cd web && npm install && npm run build && cp -r dist/* ../static/ 2>/dev/null || true
cd ..
go run ./cmd/server --bind-address=:8080
```

**Option 2 — frontend dev with hot reload:**

```bash
# Terminal 1: backend
cd dataflow-web && go run ./cmd/server --bind-address=:8080

# Terminal 2: frontend (proxy to :8080)
cd dataflow-web/web && npm install && npm run dev
```

Open http://localhost:8080 (option 1) or http://localhost:5173 (option 2). Cluster access requires `KUBECONFIG` or in-cluster config.

## Deploying to cluster

Deploy via Helm with `gui.enabled=true`. See [Web GUI documentation](../docs/docs/en/gui.md) for configuration and deployment details.

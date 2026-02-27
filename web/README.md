# DataFlow Web UI (Vue 3 + Vite)

Frontend for the DataFlow management panel.

## Development

```bash
npm install
npm run dev
```

The app runs at http://localhost:5173. API requests are proxied to the backend (default http://localhost:8080). Run the Go server separately:

```bash
# from repo root
go run ./cmd/server --bind-address=:8080
```

## Build

```bash
npm run build
```

Output goes to `web/dist/`. The Docker image builds this in a separate stage and copies it to `static/`.

## Tests

```bash
npm run test        # watch mode
npm run test:run    # single run
```

## Structure

- `src/api/` — API client
- `src/components/` — reusable components
- `src/views/` — pages (Dashboard, Manifests, Logs, Metrics)
- `src/router/` — Vue Router routes
- `src/composables/` — composables (toasts, etc.)

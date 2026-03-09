# Stage 1: build frontend (Vue + Vite)
# VERSION is passed at build time for image labels (semver from CI)
ARG VERSION=dev
FROM node:22-alpine AS frontend
WORKDIR /workspace/web
COPY web/package.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: build Go server and pack static
FROM golang:1.25-alpine AS builder
WORKDIR /workspace

RUN apk add --no-cache git

COPY go.* ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY --from=frontend /workspace/web/dist ./static

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o server ./cmd/server

# Final stage
FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /workspace/server .
COPY --from=builder /workspace/static /app/static
EXPOSE 8080
CMD ["./server", "--bind-address=:8080"]

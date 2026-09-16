# Multi-stage build for apps/base: build the web frontend, build the Go base
# server, then run base serving both the API and the static SPA on $PORT.
#
# Build context must be the monorepo root:
#   docker build -f apps/base/Dockerfile -t base .
# On Railway, set RAILWAY_DOCKERFILE_PATH=apps/base/Dockerfile on the service.

# ---- Stage 1: build the web frontend ----
FROM node:24-slim AS web
ENV CI=true
# corepack reads packageManager from the root package.json, so the pnpm
# version stays in one place rather than being pinned again here.
RUN corepack enable
WORKDIR /repo
COPY . .
RUN corepack install && pnpm install --frozen-lockfile --filter @thom/web... --filter @thom/ui...

# The SPA and the API are served from the same origin, so the client needs no
# absolute base URL. Overridable for a split deployment.
ARG VITE_API_URL=""
ENV VITE_API_URL=${VITE_API_URL}
RUN pnpm --filter @thom/web build


# ---- Stage 2: build the Go server ----
FROM golang:1.27 AS base
WORKDIR /src
# folio-core is a replace target of apps/base, so it has to be present before
# `go mod download` can resolve the module graph.
COPY packages/folio-core/ packages/folio-core/
COPY apps/base/go.mod apps/base/go.sum apps/base/
WORKDIR /src/apps/base
RUN go mod download
WORKDIR /src
COPY apps/base/ apps/base/
WORKDIR /src/apps/base
RUN CGO_ENABLED=0 go build -o /out/base .


# ---- Stage 3: runtime ----
FROM debian:bookworm-slim
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=base /out/base ./base
# publicDir() resolves pb_public next to the binary, so no env var is needed.
COPY --from=web /repo/apps/web/dist ./pb_public

# The folio collections install from the OnBootstrap hook that serve runs, so
# there is no separate migrate step. Railway injects $PORT; pb_data is a
# mounted volume.
CMD ["sh", "-c", "exec ./base serve --http=0.0.0.0:${PORT:-8080} --dir=/app/pb_data"]

# syntax=docker/dockerfile:1

########################################
# Stage 1: Build the Go backend binary
########################################
FROM golang:1.25-alpine AS backend-builder

WORKDIR /src/backend

COPY backend/go.mod ./
RUN GOSUMDB=off go mod download

COPY backend/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/resume-generator .

########################################
# Stage 2: Build the Next.js frontend
########################################
FROM node:20-alpine AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/. .

# Runtime API calls are proxied through the Next.js server to the backend
# process running alongside it in the same container (see next.config.mjs).
ENV BACKEND_API_URL=http://127.0.0.1:8080

RUN npm run build

########################################
# Stage 3: Final runtime image (backend + frontend together)
########################################
FROM node:20-alpine

ENV TECTONIC_VERSION=0.16.9
ENV XDG_CACHE_HOME=/var/cache

# Install Tectonic (LaTeX engine) + fonts + CA certs, matching backend/Dockerfile
RUN apk add --no-cache \
    ca-certificates \
    curl \
    tar \
    fontconfig \
    ttf-dejavu \
    bash \
    && fc-cache -fv \
    && curl -fsSL "https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40${TECTONIC_VERSION}/tectonic-${TECTONIC_VERSION}-x86_64-unknown-linux-musl.tar.gz" -o /tmp/tectonic.tar.gz \
    && tar -xzf /tmp/tectonic.tar.gz -C /usr/local/bin tectonic \
    && chmod +x /usr/local/bin/tectonic \
    && rm /tmp/tectonic.tar.gz \
    && mkdir -p "/var/cache/Tectonic" /tmp/tectonic-warmup

# ---------- Backend ----------
WORKDIR /app/backend

COPY --from=backend-builder /out/resume-generator /usr/local/bin/resume-generator
COPY backend/templates ./templates
COPY backend/resume.json ./

# Generate the real LaTeX document (used to warm the Tectonic cache below)
RUN resume-generator \
    -input resume.json \
    -output /tmp/tectonic-warmup \
    -format latex

# Warm the Tectonic cache (downloads fonts and packages into XDG_CACHE_HOME/Tectonic)
RUN set -eux; \
    if tectonic \
    --outdir /tmp/tectonic-warmup \
    /tmp/tectonic-warmup/resume.tex; then \
    echo "Tectonic cache warmup succeeded."; \
    else \
    echo "WARNING: Tectonic cache warmup failed (likely DNS/network)." >&2; \
    echo "WARNING: First runtime PDF generation will require network access." >&2; \
    fi

RUN rm -rf /tmp/tectonic-warmup

# ---------- Frontend ----------
WORKDIR /app/frontend

COPY --from=frontend-builder /src/frontend/.next ./.next
COPY --from=frontend-builder /src/frontend/node_modules ./node_modules
COPY --from=frontend-builder /src/frontend/package.json ./package.json
COPY --from=frontend-builder /src/frontend/next.config.mjs ./next.config.mjs

# ---------- Entrypoint ----------
WORKDIR /app
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENV BACKEND_API_URL=http://127.0.0.1:8080
ENV BACKEND_PORT=:8080
ENV FRONTEND_PORT=3000

# The SQLite database lives at /app/backend/data/resumes.db by default and is
# ephemeral unless this path is mounted to a host volume at runtime, e.g.:
#   docker run -v $(pwd)/data:/app/backend/data ...
EXPOSE 3000

ENTRYPOINT ["/app/docker-entrypoint.sh"]
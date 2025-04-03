# Node dependencies stage
FROM public.ecr.aws/docker/library/node:lts-alpine AS frontend-dependencies
WORKDIR /app

# Install pnpm globally (caching layer)
RUN npm install -g pnpm

# Copy package.json and lockfile to leverage caching
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --shamefully-hoist

# Build Nuxt (frontend) stage
FROM public.ecr.aws/docker/library/node:lts-alpine AS frontend-builder
WORKDIR /app

# Install pnpm globally again (it can reuse the cache if not changed)
RUN npm install -g pnpm

# Copy over source files and node_modules from dependencies stage
COPY frontend . 
COPY --from=frontend-dependencies /app/node_modules ./node_modules
RUN pnpm build

# Go dependencies stage
FROM public.ecr.aws/docker/library/golang:alpine AS builder-dependencies
WORKDIR /go/src/app

# Copy go.mod and go.sum for better caching
COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download

# Build API stage
FROM public.ecr.aws/docker/library/golang:alpine AS builder
ARG BUILD_TIME
ARG COMMIT
ARG VERSION

# Install necessary build tools
RUN apk update && \
    apk upgrade && \
    apk add --no-cache git build-base gcc g++

WORKDIR /go/src/app

# Copy Go modules (from dependencies stage) and source code
COPY --from=builder-dependencies /go/pkg/mod /go/pkg/mod
COPY ./backend .

# Clear old public files and copy new ones from frontend build
RUN rm -rf ./app/api/public
COPY --from=frontend-builder /app/.output/public ./app/api/static/public

# Use cache for Go build artifacts
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME -X main.version=$VERSION" \
    -o /go/bin/api \
    -v ./app/api/*.go

# Production stage
FROM public.ecr.aws/docker/library/alpine:latest
ENV HBOX_MODE=production
ENV HBOX_STORAGE_DATA=/data/
ENV HBOX_DATABASE_SQLITE_PATH=/data/homebox.db?_pragma=busy_timeout=2000&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates wget

# Create application directory and copy over built Go binary
RUN mkdir /app
COPY --from=builder /go/bin/api /app
RUN chmod +x /app/api

# Labels and configuration for the final image
LABEL Name=homebox Version=0.0.1
LABEL org.opencontainers.image.source="https://github.com/sysadminsmedia/homebox"

# Expose necessary ports for Homebox
EXPOSE 7745
WORKDIR /app

# Healthcheck configuration
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD [ "wget", "--no-verbose", "--tries=1", "-O", "-", "http://localhost:7745/api/v1/status" ]

# Persist volume
VOLUME [ "/data" ]

# Entrypoint and CMD
ENTRYPOINT [ "/app/api" ]
CMD [ "/data/config.yml" ]

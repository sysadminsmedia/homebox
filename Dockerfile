
# Build Nuxt
FROM node:18-alpine AS frontend-builder
WORKDIR  /app
RUN npm install -g pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --shamefully-hoist
COPY frontend .
RUN pnpm build

# Build API
FROM golang:alpine AS builder
ARG BUILD_TIME
ARG COMMIT
ARG VERSION
ARG BUSYBOX_VERSION=1.36.1-r31
RUN apk update && \
    apk upgrade && \
    apk add --update git build-base gcc g++

WORKDIR /go/src/app
COPY ./backend .
RUN go get -d -v ./...
RUN rm -rf ./app/api/public
COPY --from=frontend-builder /app/.output/public ./app/api/static/public
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME -X main.version=$VERSION"  \
    -o /go/bin/api \
    -v ./app/api/*.go

# Downloading Wget
ADD https://busybox.net/downloads/binaries/$BUSYBOX_VERSION/busybox_WGET /wget
RUN chmod a+x /wget

FROM gcr.io/distroless/java:latest

COPY --from=builder /wget /usr/bin/wget

# Production Stage
FROM alpine:latest

ENV HBOX_MODE=production
ENV HBOX_STORAGE_DATA=/data/
ENV HBOX_STORAGE_SQLITE_URL=/data/homebox.db?_pragma=busy_timeout=2000&_pragma=journal_mode=WAL&_fk=1

RUN apk --no-cache add ca-certificates
RUN mkdir /app
COPY --from=builder /go/bin/api /app
COPY --from=builder /wget /usr/bin/wget

RUN chmod +x /app/api

LABEL Name=homebox Version=0.0.1
LABEL org.opencontainers.image.source="https://github.com/sysadminsmedia/homebox"
EXPOSE 7745
WORKDIR /app
HEALTHCHECK --interval=30s \
            --timeout=5s \
            --start-period=5s \
            --retries=3 \
            CMD [ "/usr/bin/wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:7745/api/v1/status" ]
VOLUME [ "/data" ]

ENTRYPOINT [ "/app/api" ]
CMD [ "/data/config.yml" ]

# syntax=docker/dockerfile:1

FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal ./internal
COPY pkg ./pkg
COPY cmd/telegram-ingestor ./cmd/telegram-ingestor
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/video-server . \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/telegram-ingestor ./cmd/telegram-ingestor

FROM debian:bookworm-slim

ARG APP_UID=10001
ARG APP_GID=10001

# Debian's ffmpeg package provides both ffmpeg and ffprobe.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg \
    && ffmpeg -version >/dev/null \
    && ffprobe -version >/dev/null \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid "${APP_GID}" video \
    && useradd --uid "${APP_UID}" --gid "${APP_GID}" --create-home --home-dir /home/video --shell /usr/sbin/nologin video \
    && mkdir -p /data/storage /data/tmp/uploads /data/telegram-session \
    && chown -R "${APP_UID}:${APP_GID}" /data

WORKDIR /app
COPY --from=builder --chown=video:video /out/video-server /app/video-server
COPY --from=builder --chown=video:video /out/telegram-ingestor /app/telegram-ingestor

USER video:video
CMD ["/app/video-server"]

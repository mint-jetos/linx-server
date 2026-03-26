#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend
WORKDIR /app

RUN corepack enable

COPY assets/static/package.json assets/static/pnpm-*.yaml .
RUN --mount=type=cache,target=/root/.cache \
  pnpm install --frozen-lockfile

COPY assets/static .
RUN --mount=type=cache,target=/root/.cache \
  pnpm run build

FROM --platform=$BUILDPLATFORM golang:1.26.0-alpine AS backend
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=frontend /app/dist assets/static/dist

ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
  go build -ldflags='-w -s' -trimpath

FROM alpine:3.23.3
WORKDIR /data

# --userns="keep-id:uid=1000,gid=1000" -v /path/to/data:/data

RUN adduser -D -u 1000 ubuntu

COPY --from=backend /app/linx-server /usr/bin

RUN <<EOT
  set -eux
  mkdir -p /data/files
  mkdir -p /data/meta
  chown -R 1000:1000 /data
EOT

VOLUME "/data"

EXPOSE 8080

# Use the name you defined above or just the number 1000
USER ubuntu
ENV LINX_DEFAULTS=container
ENV LINX_CONFIG=/data/config.toml
ENTRYPOINT ["/usr/bin/linx-server"]

# Multi-stage build for the BillionMail core image.
# Build context MUST be the repository root:
#   docker build -f Dockerfiles/core/Dockerfile -t billionmail-core .

# ---------- Stage 1: frontend ----------
FROM node:22-bookworm-slim AS frontend

WORKDIR /src/core/frontend
RUN corepack enable && corepack prepare pnpm@12 --activate

COPY core/frontend/package.json core/frontend/pnpm-lock.yaml core/frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY core/frontend/ ./
RUN pnpm build

# ---------- Stage 2: backend ----------
FROM golang:1.24-alpine AS backend

WORKDIR /src/core
COPY core/go.mod core/go.sum ./
RUN go mod download

COPY core/ ./
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/billionmail main.go

# ---------- Stage 3: runtime ----------
FROM alpine:3.20

LABEL org.opencontainers.image.source="https://github.com/tuannv2910/UTC-BillionMail"

# Set environment variables
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Install dependencies (using Alpine's apk package manager)
RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    supervisor \
    rsyslog \
    tzdata \
    busybox-extras \
    postgresql-client \
    fail2ban \
    iptables \
    ipset \
    && rm -rf /var/cache/apk/* \
    && addgroup -S crontab \
    && addgroup -S www && adduser -S -D -H -G www www

COPY --from=backend /out/billionmail /opt/billionmail/core/billionmail
COPY core/manifest /opt/billionmail/core/manifest
COPY core/languages /opt/billionmail/core/languages
COPY core/public /opt/billionmail/core/public
COPY --from=frontend /src/core/frontend/dist /opt/billionmail/core/public/dist
COPY core/resource /opt/billionmail/core/resource
COPY core/template /opt/billionmail/core/template

COPY Dockerfiles/core/stop-supervisor.sh /stop-supervisor.sh
COPY Dockerfiles/core/core.sh /core.sh
COPY Dockerfiles/core/restart_fail2ban.sh /restart_fail2ban.sh
COPY Dockerfiles/core/supervisord.conf /etc/supervisor/supervisord.conf
COPY Dockerfiles/core/fail2ban.conf /etc/fail2ban/fail2ban.conf

# Strip CRLF in case the image is built from a Windows checkout
RUN sed -i 's/\r$//' /stop-supervisor.sh /core.sh /restart_fail2ban.sh \
        /etc/supervisor/supervisord.conf /etc/fail2ban/fail2ban.conf \
    && chmod +x /stop-supervisor.sh /core.sh /restart_fail2ban.sh /opt/billionmail/core/billionmail

ENTRYPOINT ["/core.sh"]
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]

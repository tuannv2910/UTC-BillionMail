# CI/CD

Workflow: `.github/workflows/ci.yml`. Image: `ghcr.io/tuannv2910/billionmail-core`.

## Pipeline

| Trigger            | backend + frontend tests | build image | push to GHCR | tags                         |
|--------------------|:------------------------:|:-----------:|:------------:|------------------------------|
| Pull request → main| ✅                       | ✅          | ❌           | —                            |
| Push to main       | ✅                       | ✅          | ✅           | `latest`, `sha-<short>`      |
| Push tag `v1.2.3`  | ✅                       | ✅          | ✅           | `1.2.3`, `1.2`, `sha-<short>`|

Jobs:
1. **backend** — `gofmt`, `go vet ./...`, `go test -short ./internal/service/...`
2. **frontend** — `pnpm lint`, `pnpm test`
3. **image** — runs only if both pass; builds `Dockerfiles/core/Dockerfile` (multi-stage: Node → Go → Alpine runtime).

Only the `core` service is custom. Postgres, Redis, Rspamd, Dovecot, Postfix and Roundcube use upstream images;
their config lives in `conf/` and is bind-mounted at runtime.

## Release a version

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Build locally (same as CI)

```bash
docker build -f Dockerfiles/core/Dockerfile -t billionmail-core:local .
```

The build context is the repo root; `.dockerignore` is an allowlist so runtime data dirs are never sent.

## Use the image

`docker-compose.yml` reads the core image from `CORE_IMAGE` (default: upstream `billionmail/core:4.9.3`).
In `.env`:

```env
CORE_IMAGE=ghcr.io/tuannv2910/billionmail-core:1.0.0
```

Then `docker compose pull core-billionmail && docker compose up -d core-billionmail`.
Roll back by setting the previous tag.

## GHCR visibility

The first push creates a **private** package. To let others pull without login:
GitHub → Profile → Packages → `billionmail-core` → Package settings → Change visibility → Public.
Otherwise run `docker login ghcr.io` on the server with a PAT that has `read:packages`.

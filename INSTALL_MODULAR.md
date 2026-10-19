# BillionMail Modular Installer

Instead of running the monolithic `install.sh`, you can run each phase separately
to make troubleshooting and maintenance easier.

## File map

| File                   | Role                                                                     |
|------------------------|--------------------------------------------------------------------------|
| `install-lib.sh`       | Shared library: constants, logging, OS detection, state save/load        |
| `install-docker.sh`    | Docker-specific library: mirror selection, Docker CE install, Compose    |
| `01-env-prepare.sh`    | **Step 1** — Host prep: packages, secrets, timezone, backup old config   |
| `02-docker-install.sh` | **Step 2** — Install Docker + Compose, start `docker.service`            |
| `03-service-deploy.sh` | **Step 3** — Write `billionmail.conf` + `.env`, SSL cert, compose up -d  |
| `04-service-init.sh`   | **Step 4** — Import schema, create domain, DKIM, mailbox, print DNS info |

## How state flows between steps

```
01-env-prepare.sh  ──writes──►  install-state.env  (secrets, hostname, TZ, ports)
        │
        ▼
02-docker-install.sh  (no shared state — installs Docker)
        │
        ▼
03-service-deploy.sh  ──reads──►  install-state.env
                       ──writes──►  billionmail.conf
                       ──writes──►  .env   (same content, used by docker-compose)
                       ──writes──►  ssl-self-signed/{key,cert}.pem
        │
        ▼
04-service-init.sh   ──reads──►  billionmail.conf  (prefers this over state file)
                       ──writes──►  install-mailbox-info.txt (mailbox creds)
```

If `billionmail.conf` already exists when step 4 runs, step 4 sources that file
instead of `install-state.env`. So state always flows forward, never backward.

## Usage (run in order, as root)

```bash
# Step 1: prepare environment
sudo bash 01-env-prepare.sh --domain mail.example.com --TZ Asia/Ho_Chi_Minh

# Step 2: install Docker (idempotent)
sudo bash 02-docker-install.sh

# Step 3: deploy containers
sudo bash 03-service-deploy.sh

# Step 4 (optional): init DB + create domain + mailbox + DKIM
sudo bash 04-service-init.sh
```

## Idempotency

| Script                 | Re-run safe? | What it does on re-run                                          |
|------------------------|--------------|-----------------------------------------------------------------|
| `01-env-prepare.sh`    | ✅           | Reuses `DBPASS_file.pl`; re-generates only SafePath/etc         |
| `02-docker-install.sh` | ✅           | Detects existing Docker, skips install                          |
| `03-service-deploy.sh` | ⚠️           | Asks confirmation; regenerates config + SSL cert                |
| `04-service-init.sh`   | ✅           | `INSERT ... ON CONFLICT` style — skips if domain/mailbox exists |

## Backwards compatibility

The original `install.sh` still works exactly as before — nothing has been removed
or renamed in it. The new scripts are an alternative for incremental maintenance.

## Maintenance tips

- To **change host packages** → edit `install-lib.sh` → `Install_RPM_Pack` /
  `Install_Deb_Pack` functions.
- To **change Docker repo selection logic** → edit `install-docker.sh` →
  `Docker_Download_Url_Check`.
- To **change mail/HTTP/HTTPS port defaults** → edit `install-lib.sh` →
  constants block at the top.
- To **change DB schema import / domain creation logic** → edit
  `04-service-init.sh`.
- To **add a new service config to `billionmail.conf`** → edit
  `03-service-deploy.sh` (the `cat <<EOF` block).

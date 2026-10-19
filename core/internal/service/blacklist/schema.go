package blacklist

// SchemaSQL creates the suppression list table and its indexes. Every statement is idempotent.
var SchemaSQL = []string{
	`CREATE TABLE IF NOT EXISTS bm_email_blacklist (
		id             BIGSERIAL PRIMARY KEY,
		scope          SMALLINT     NOT NULL DEFAULT 1,   -- 1: email, 2: domain
		value          VARCHAR(320) NOT NULL,             -- normalized email or domain
		domain         VARCHAR(255) NOT NULL DEFAULT '',
		reason         SMALLINT     NOT NULL,
		source         SMALLINT     NOT NULL,
		status         SMALLINT     NOT NULL DEFAULT 0,   -- 0: observing, 1: active, 2: released
		hit_count      INTEGER      NOT NULL DEFAULT 1,
		dsn            VARCHAR(16)  NOT NULL DEFAULT '',
		last_response  TEXT         NOT NULL DEFAULT '',
		last_task_id   INTEGER      NOT NULL DEFAULT 0,
		expires_at     BIGINT       NOT NULL DEFAULT 0,   -- 0: permanent
		note           VARCHAR(255) NOT NULL DEFAULT '',
		first_seen_at  BIGINT       NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
		last_seen_at   BIGINT       NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
		create_time    BIGINT       NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
		update_time    BIGINT       NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
		UNIQUE (scope, value)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_bm_email_blacklist_status_reason ON bm_email_blacklist (status, reason)`,
	`CREATE INDEX IF NOT EXISTS idx_bm_email_blacklist_domain ON bm_email_blacklist (domain)`,
	`CREATE INDEX IF NOT EXISTS idx_bm_email_blacklist_last_seen ON bm_email_blacklist (last_seen_at)`,
	`CREATE INDEX IF NOT EXISTS idx_bm_email_blacklist_expires ON bm_email_blacklist (expires_at) WHERE expires_at > 0 AND status = 1`,
}

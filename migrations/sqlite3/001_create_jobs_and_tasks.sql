-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
PRAGMA foreign_keys = ON;

CREATE TABLE jobs (
    id BLOB PRIMARY KEY,
    name TEXT NOT NULL,
	description TEXT,
	config_id BLOB NOT NULL,
	config_version BLOB NOT NULL,
	state TEXT NOT NULL,
	progress REAL NOT NULL DEFAULT 0.0,
	submit_date TEXT NOT NULL,
    start_date  TEXT,
    end_date    TEXT,

	FOREIGN KEY(config_id, config_version) REFERENCES job_configs(id, version) ON DELETE RESTRICT
);

CREATE TABLE job_configs (
	id BLOB NOT NULL,
    version BLOB NOT NULL,
    name TEXT NOT NULL,
	description TEXT,
	details TEXT NOT NULL,

	is_default INTEGER NOT NULL DEFAULT 0,
	default_key INTEGER GENERATED ALWAYS AS (CASE WHEN is_default THEN 1 ELSE NULL END) VIRTUAL,
	UNIQUE(default_key),

	PRIMARY KEY (id, version)
);

CREATE TABLE task_runs (
	id BLOB PRIMARY KEY,
	job_id BLOB NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	task_name TEXT NOT NULL,
	state TEXT NOT NULL,
	start_date TEXT,
	end_date TEXT,
	details TEXT,

	FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE INDEX idx_jobs_state ON jobs(state);
CREATE INDEX idx_task_runs_job_id ON task_runs(job_id);
CREATE INDEX idx_task_runs_state ON task_runs(state);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS job_configs;
DROP TABLE IF EXISTS task_runs;

DROP INDEX IF EXISTS idx_jobs_state;
DROP INDEX IF EXISTS idx_task_runs_job_id;
DROP INDEX IF EXISTS idx_task_runs_state;

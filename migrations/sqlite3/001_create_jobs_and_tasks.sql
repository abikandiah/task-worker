-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
PRAGMA foreign_keys = ON;

CREATE TABLE jobs (
    id BLOB PRIMARY KEY,
    name TEXT NOT NULL,
	description TEXT,
	config_id BLOB NOT NULL,
state TEXT NOT NULL,
	progress REAL NOT NULL DEFAULT 0.0,
	submit_date TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now')),
    start_date  TEXT,
    end_date    TEXT,

	FOREIGN KEY(config_id) REFERENCES job_configs(id) ON DELETE RESTRICT
);

CREATE TABLE job_configs (
	id BLOB KEY,
    version TEXT NOT NULL,
    name TEXT NOT NULL,
	details TEXT NOT NULL,

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

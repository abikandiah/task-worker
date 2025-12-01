-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE job_configs (
    id UUID NOT NULL,
    version UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    details TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id, version)
);

-- Create a unique partial index to enforce only one default config
CREATE UNIQUE INDEX idx_job_configs_default 
    ON job_configs(is_default) 
    WHERE is_default = TRUE;

CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    config_id UUID NOT NULL,
    config_version UUID NOT NULL,
    state TEXT NOT NULL,
    progress REAL NOT NULL DEFAULT 0.0,
    submit_date TIMESTAMP NOT NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    FOREIGN KEY(config_id, config_version) 
        REFERENCES job_configs(id, version) 
        ON DELETE RESTRICT
);

CREATE TABLE task_runs (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    task_name TEXT NOT NULL,
    state TEXT NOT NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    details TEXT,
    FOREIGN KEY(job_id) 
        REFERENCES jobs(id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_jobs_state ON jobs(state);
CREATE INDEX idx_task_runs_job_id ON task_runs(job_id);
CREATE INDEX idx_task_runs_state ON task_runs(state);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

DROP INDEX IF EXISTS idx_task_runs_state;
DROP INDEX IF EXISTS idx_task_runs_job_id;
DROP INDEX IF EXISTS idx_jobs_state;
DROP INDEX IF EXISTS idx_job_configs_default;

DROP TABLE IF EXISTS task_runs;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS job_configs;
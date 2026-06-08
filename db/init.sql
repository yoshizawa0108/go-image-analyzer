CREATE TABLE IF NOT EXISTS analysis_results (
    id         SERIAL PRIMARY KEY,
    filename   TEXT,
    objects    JSONB        NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);
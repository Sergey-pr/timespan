-- migrate:up
CREATE TABLE IF NOT EXISTS tasks (
    id          INTEGER PRIMARY KEY,
    title       TEXT    NOT NULL,
    description TEXT,
    status      TEXT    NOT NULL DEFAULT 'pending',
    elapsed_ms  INTEGER NOT NULL DEFAULT 0,
    started_at  DATETIME,
    created_at  DATETIME NOT NULL,
    finished_at DATETIME
);

-- migrate:down
DROP TABLE IF EXISTS tasks;

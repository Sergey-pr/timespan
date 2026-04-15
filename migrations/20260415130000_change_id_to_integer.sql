-- migrate:up
-- Recreates tasks table with INTEGER PRIMARY KEY (was TEXT).
-- Existing data is preserved; new integer IDs are assigned by SQLite.
CREATE TABLE tasks_new (
    id          INTEGER PRIMARY KEY,
    title       TEXT    NOT NULL,
    description TEXT,
    status      TEXT    NOT NULL DEFAULT 'pending',
    elapsed_ms  INTEGER NOT NULL DEFAULT 0,
    started_at  DATETIME,
    created_at  DATETIME NOT NULL,
    finished_at DATETIME
);
INSERT INTO tasks_new (title, description, status, elapsed_ms, started_at, created_at, finished_at)
    SELECT title, description, status, elapsed_ms, started_at, created_at, finished_at FROM tasks;
DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

-- migrate:down
CREATE TABLE tasks_old (
    id          TEXT PRIMARY KEY,
    title       TEXT    NOT NULL,
    description TEXT,
    status      TEXT    NOT NULL DEFAULT 'pending',
    elapsed_ms  INTEGER NOT NULL DEFAULT 0,
    started_at  DATETIME,
    created_at  DATETIME NOT NULL,
    finished_at DATETIME
);
INSERT INTO tasks_old (id, title, description, status, elapsed_ms, started_at, created_at, finished_at)
    SELECT CAST(id AS TEXT), title, description, status, elapsed_ms, started_at, created_at, finished_at FROM tasks;
DROP TABLE tasks;
ALTER TABLE tasks_old RENAME TO tasks;

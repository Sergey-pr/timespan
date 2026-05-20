-- migrate:up
CREATE TABLE IF NOT EXISTS categories (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,
    created_at DATETIME NOT NULL
);

ALTER TABLE tasks ADD COLUMN category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL;

-- migrate:down
ALTER TABLE tasks DROP COLUMN category_id;
DROP TABLE IF EXISTS categories;

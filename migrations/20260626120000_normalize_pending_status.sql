-- migrate:up
-- Earlier builds let tasks fall back to the legacy column default 'pending',
-- which is not a canonical TaskStatus and leaves such tasks unusable in the UI.
-- Normalise any survivors to the canonical initial status.
UPDATE tasks SET status = 'ready_to_start' WHERE status = 'pending';

-- migrate:down
UPDATE tasks SET status = 'pending' WHERE status = 'ready_to_start';

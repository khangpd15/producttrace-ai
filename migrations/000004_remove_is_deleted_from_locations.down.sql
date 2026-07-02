BEGIN;

ALTER TABLE locations ADD COLUMN IF NOT EXISTS is_deleted boolean DEFAULT false;

COMMIT;

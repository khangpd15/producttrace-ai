ALTER TABLE batches
DROP CONSTRAINT IF EXISTS chk_batches_status;

ALTER TABLE batches
ADD CONSTRAINT chk_batches_status
CHECK (status IN (
    'ACTIVE',
    'CLOSED',
    'RECALLED',
    'BLOCKED'
));
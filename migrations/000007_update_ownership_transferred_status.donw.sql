ALTER TABLE ownerships
DROP CONSTRAINT IF EXISTS chk_ownerships_status;

ALTER TABLE ownerships
ADD CONSTRAINT chk_ownerships_status
CHECK (status IN (
    'PENDING',
    'ACTIVE',
    'REVOKED',
    'TRANSFERRED'
));
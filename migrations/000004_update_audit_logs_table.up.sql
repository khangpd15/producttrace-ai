ALTER TABLE "auditLog" RENAME TO audit_logs;

ALTER TABLE audit_logs
    DROP COLUMN content,
    DROP COLUMN type;

ALTER TABLE audit_logs
    ADD COLUMN user_id UUID,
    ADD COLUMN action VARCHAR(20),
    ADD COLUMN entity VARCHAR(50),
    ADD COLUMN entity_id UUID,
    ADD COLUMN old_data JSONB,
    ADD COLUMN new_data JSONB;

ALTER TABLE audit_logs
    ALTER COLUMN created_at SET DEFAULT NOW();
ALTER TABLE audit_logs
    DROP COLUMN user_id,
    DROP COLUMN action,
    DROP COLUMN entity,
    DROP COLUMN entity_id,
    DROP COLUMN old_data,
    DROP COLUMN new_data;

ALTER TABLE audit_logs
    ADD COLUMN content TEXT NOT NULL DEFAULT '',
    ADD COLUMN type VARCHAR;

ALTER TABLE audit_logs
    ALTER COLUMN created_at DROP DEFAULT;

ALTER TABLE audit_logs
    RENAME TO "auditLog";
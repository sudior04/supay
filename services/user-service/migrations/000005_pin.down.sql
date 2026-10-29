DROP TABLE pin_audit_events;
DROP TABLE auth_challenges;
ALTER TABLE refresh_sessions DROP COLUMN pin_verified;
ALTER TABLE user_credentials DROP COLUMN pin_hash, DROP COLUMN pin_failures, DROP COLUMN pin_locked_until;

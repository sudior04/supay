ALTER TABLE user_credentials
    ADD COLUMN pin_hash text,
    ADD COLUMN pin_failures integer NOT NULL DEFAULT 0 CHECK (pin_failures BETWEEN 0 AND 5),
    ADD COLUMN pin_locked_until timestamptz;
ALTER TABLE refresh_sessions ADD COLUMN pin_verified boolean NOT NULL DEFAULT false;
UPDATE refresh_sessions SET revoked_at=now() WHERE revoked_at IS NULL;

CREATE TABLE auth_challenges (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    purpose text NOT NULL CHECK (purpose IN ('LOGIN','PIN_SETUP')),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts BETWEEN 0 AND 5),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX auth_challenges_user_idx ON auth_challenges(user_id);
CREATE TABLE pin_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid NOT NULL REFERENCES users(id),
    target_user_id uuid NOT NULL REFERENCES users(id),
    action text NOT NULL CHECK (action IN ('ENROLL','CHANGE','ADMIN_RESET')),
    created_at timestamptz NOT NULL DEFAULT now()
);

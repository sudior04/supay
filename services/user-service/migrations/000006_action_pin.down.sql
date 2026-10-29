ALTER TABLE refresh_sessions ADD COLUMN pin_verified boolean NOT NULL DEFAULT false;
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

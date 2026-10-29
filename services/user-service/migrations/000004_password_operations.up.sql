CREATE TABLE password_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid NOT NULL REFERENCES users(id),
    target_user_id uuid NOT NULL REFERENCES users(id),
    action text NOT NULL CHECK (action IN ('SELF_CHANGE','ADMIN_RESET')),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX password_audit_events_target_created_idx ON password_audit_events(target_user_id, created_at DESC);

ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('CUSTOMER','MERCHANT','ADMIN'));
ALTER TABLE users ADD COLUMN is_demo boolean NOT NULL DEFAULT false;

INSERT INTO users (name, email, role, is_demo) VALUES
    ('Supay Customer', 'customer@gmail.com', 'CUSTOMER', true),
    ('Supay Merchant', 'merchant@gmail.com', 'MERCHANT', true),
    ('Supay Admin', 'admin@gmail.com', 'ADMIN', true)
ON CONFLICT (lower(email)) DO UPDATE
    SET name = EXCLUDED.name, role = EXCLUDED.role, is_demo = true;

INSERT INTO user_credentials (user_id, password_hash)
SELECT id, '$2a$10$iKMBBXCxBfzDf3opsXWvy.ULzjHdwTk33pdYSlaaCYqL2xV8yduEW'
FROM users WHERE lower(email) = 'customer@gmail.com'
ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, password_changed_at = now();

INSERT INTO user_credentials (user_id, password_hash)
SELECT id, '$2a$10$c0YbJtv5ijfEFpeJ48Ok5eeDzawI7H6IvIGlkKck2kjucWNrX0a1G'
FROM users WHERE lower(email) = 'merchant@gmail.com'
ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, password_changed_at = now();

INSERT INTO user_credentials (user_id, password_hash)
SELECT id, '$2a$10$nrngEktAffGOgEQR.PEopexG.j/2eIux/G9VKeQwnobwYERpbdYmS'
FROM users WHERE lower(email) = 'admin@gmail.com'
ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, password_changed_at = now();

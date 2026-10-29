-- PIN now confirms individual business mutations; login/refresh only require password.
DROP TABLE auth_challenges;
ALTER TABLE refresh_sessions DROP COLUMN pin_verified;

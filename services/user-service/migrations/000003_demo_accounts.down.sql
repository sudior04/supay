DELETE FROM users WHERE is_demo = true AND lower(email) IN ('customer@gmail.com','merchant@gmail.com','admin@gmail.com');
ALTER TABLE users DROP COLUMN is_demo;
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('CUSTOMER','MERCHANT'));

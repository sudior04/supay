-- Migration plumbing only; domain tables come with use cases.
CREATE TABLE service_metadata (
 singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
 service_name text NOT NULL,
 initialized_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO service_metadata (service_name) VALUES ('wallet-service');

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id bigserial Primary KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated  bool NOT NULL,
    version integer NOT NULL DEFAULT 1


)
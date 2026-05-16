CREATE TABLE IF NOT EXISTS energy_document (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    about text NOT NULL,
    storage_url text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);
-- =====================================================================
-- Renewable Buddy: lean PostgreSQL + TimescaleDB schema (CV project)
--
-- Assumes a `users` table with a BIGINT id already exists (your auth
-- tables). For a fresh dev database, run scripts/dev_users_stub.sql first.
--
-- Apply with psql, which runs each statement on its own. If you switch to
-- a migration tool, put the continuous aggregates (section 5) in their own
-- migration with transactions turned off: CREATE MATERIALIZED VIEW ...
-- WITH (timescaledb.continuous) cannot run inside BEGIN/COMMIT.
-- =====================================================================

CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ---------------------------------------------------------------------
-- 0. Helpers
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------
-- 1. Tariffs (Egyptian residential pricing is tiered and changes often,
--    so keep the rates in data, not in code)
-- ---------------------------------------------------------------------
CREATE TABLE tariff_schemes (
    id             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title           VARCHAR(100) NOT NULL,
    currency       VARCHAR(10) NOT NULL DEFAULT 'EGP',
    effective_from DATE NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tariff_tiers (
    scheme_id    INT NOT NULL REFERENCES tariff_schemes(id) ON DELETE CASCADE,
    tier_no      SMALLINT NOT NULL,
    from_kwh     NUMERIC(10, 2) NOT NULL,            -- inclusive lower bound
    to_kwh       NUMERIC(10, 2),                     -- NULL = no upper bound
    rate_per_kwh NUMERIC(10, 4) NOT NULL CHECK (rate_per_kwh >= 0),
    PRIMARY KEY (scheme_id, tier_no),
    CHECK (to_kwh IS NULL OR to_kwh > from_kwh)
);
-- Load current official rates yourself; they changed twice in 2026.

-- ---------------------------------------------------------------------
-- 2. Projects
-- ---------------------------------------------------------------------
CREATE TABLE projects (
    id                    INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id               INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title                  VARCHAR(255) NOT NULL,
    monthly_bill          NUMERIC(12, 2) NOT NULL CHECK (monthly_bill >= 0),

    -- Use a tiered scheme when known, otherwise a flat fallback rate
    tariff_scheme_id      INT REFERENCES tariff_schemes(id),
    flat_electricity_rate NUMERIC(10, 4) CHECK (flat_electricity_rate >= 0),

    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (tariff_scheme_id IS NOT NULL OR flat_electricity_rate IS NOT NULL)
);

CREATE INDEX idx_projects_user ON projects (user_id);

CREATE TRIGGER trg_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------
-- 3. Devices
-- ---------------------------------------------------------------------
CREATE TABLE devices (
    id           INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id   INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    device_uid   VARCHAR(100) NOT NULL UNIQUE,          -- MAC address or chip ID
    name         VARCHAR(100),
    api_key_hash TEXT NOT NULL,                         -- store a hash, never the raw key
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    last_seen_at TIMESTAMPTZ,                           -- update at most ~once a minute
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_project ON devices (project_id);

-- ---------------------------------------------------------------------
-- 4. Telemetry (TimescaleDB hypertable)
--    recorded_at comes from the DEVICE so buffered readings keep their
--    true time; received_at is when the server got them.
-- ---------------------------------------------------------------------
CREATE TABLE telemetry_logs (
    device_id         INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    recorded_at       TIMESTAMPTZ NOT NULL,
    received_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    pv_power_watts    REAL,
    battery_soc_pct   REAL CHECK (battery_soc_pct BETWEEN 0 AND 100),
    battery_voltage_v REAL,
    load_power_watts  REAL,
    grid_power_watts  REAL,                             -- + importing, - exporting

    PRIMARY KEY (device_id, recorded_at)
);

SELECT create_hypertable('telemetry_logs', 'recorded_at',
                         chunk_time_interval => INTERVAL '1 day');

ALTER TABLE telemetry_logs SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id',
    timescaledb.compress_orderby   = 'recorded_at DESC'
);
SELECT add_compression_policy('telemetry_logs', INTERVAL '3 days');

-- Raw data retention. The aggregates below outlive this.
SELECT add_retention_policy('telemetry_logs', INTERVAL '30 days');

-- ---------------------------------------------------------------------
-- 5. Continuous aggregates (history charts read from these)
--    Energy for a bucket = avg power x bucket length (e.g. avg_pv_watts x 1 h = Wh).
-- ---------------------------------------------------------------------

-- 5-minute rollup (today's chart)
CREATE MATERIALIZED VIEW telemetry_5min
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', recorded_at) AS bucket,
    device_id,
    avg(pv_power_watts)   AS avg_pv_watts,
    max(pv_power_watts)   AS max_pv_watts,
    avg(load_power_watts) AS avg_load_watts,
    avg(grid_power_watts) AS avg_grid_watts,
    avg(battery_soc_pct)  AS avg_battery_soc_pct,
    min(battery_soc_pct)  AS min_battery_soc_pct,
    count(*)              AS samples
FROM telemetry_logs
GROUP BY bucket, device_id
WITH NO DATA;

SELECT add_continuous_aggregate_policy('telemetry_5min',
    start_offset      => INTERVAL '2 days',
    end_offset        => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes');

SELECT add_retention_policy('telemetry_5min', INTERVAL '180 days');

-- Hourly rollup (week/month charts and daily savings; kept indefinitely)
CREATE MATERIALIZED VIEW telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', recorded_at) AS bucket,
    device_id,
    avg(pv_power_watts)   AS avg_pv_watts,
    max(pv_power_watts)   AS max_pv_watts,
    avg(load_power_watts) AS avg_load_watts,
    avg(grid_power_watts) AS avg_grid_watts,
    avg(battery_soc_pct)  AS avg_battery_soc_pct,
    min(battery_soc_pct)  AS min_battery_soc_pct,
    count(*)              AS samples
FROM telemetry_logs
GROUP BY bucket, device_id
WITH NO DATA;

-- start_offset must stay shorter than the raw retention (30 days), or a
-- refresh would try to rebuild buckets whose raw data is already dropped.
SELECT add_continuous_aggregate_policy('telemetry_hourly',
    start_offset      => INTERVAL '7 days',
    end_offset        => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
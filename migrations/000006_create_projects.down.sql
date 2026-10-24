-- Remove scheduled TimescaleDB jobs before dropping their relations.
SELECT remove_retention_policy('telemetry_5min', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('telemetry_5min', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('telemetry_hourly', if_exists => TRUE);
SELECT remove_compression_policy('telemetry_logs', if_exists => TRUE);
SELECT remove_retention_policy('telemetry_logs', if_exists => TRUE);

-- Drop aggregates before the hypertable they read from.
DROP MATERIALIZED VIEW IF EXISTS telemetry_hourly;
DROP MATERIALIZED VIEW IF EXISTS telemetry_5min;

-- Dropping the hypertable also removes its TimescaleDB chunks.
DROP TABLE IF EXISTS telemetry_logs;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS tariff_tiers;
DROP TABLE IF EXISTS tariff_schemes;

DROP FUNCTION IF EXISTS set_updated_at();

-- Intentionally leave the TimescaleDB extension installed: it may be shared
-- by other migrations or applications in this database.
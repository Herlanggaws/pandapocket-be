-- SELECT-only Postgres role for scripts/prod-read.sh (prod metrics).
-- Do NOT put a password in this file. Do not run against prod from an agent
-- unless an operator explicitly asks.
--
-- Example (superuser / peer on the deploy host):
--   sudo -u postgres psql -d pandapocket -v ON_ERROR_STOP=1 \
--     -v readonly_password='…' \
--     -f scripts/sql/create_readonly_role.sql
--
-- Then point prod-read at the role:
--   export PROD_DB_USER='pandapocket_readonly'

\if :{?readonly_password}
\else
\echo 'error: pass -v readonly_password=… (no password is stored in this file)'
\quit 1
\endif

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'pandapocket_readonly') THEN
    CREATE ROLE pandapocket_readonly WITH
      LOGIN
      NOSUPERUSER
      NOCREATEDB
      NOCREATEROLE
      INHERIT
      NOREPLICATION;
  END IF;
END
$$;

-- Password via psql variable only (never hard-coded).
ALTER ROLE pandapocket_readonly PASSWORD :'readonly_password';

ALTER ROLE pandapocket_readonly SET default_transaction_read_only = on;

GRANT CONNECT ON DATABASE pandapocket TO pandapocket_readonly;
GRANT USAGE ON SCHEMA public TO pandapocket_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO pandapocket_readonly;

-- Future tables created by the migrate/deploy role (nark) and by the
-- current session role also grant SELECT to the read-only role.
ALTER DEFAULT PRIVILEGES FOR ROLE nark IN SCHEMA public
  GRANT SELECT ON TABLES TO pandapocket_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO pandapocket_readonly;

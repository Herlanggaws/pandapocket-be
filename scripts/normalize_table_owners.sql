-- D4: Normalize public table/sequence owners to deploy role `nark`.
-- Idempotent. Run as a superuser (e.g. `sudo -u postgres psql -d pandapocket`).
--
-- Owned serial/identity sequences follow table ownership — alter tables first.
--
--   sudo -u postgres psql -d pandapocket -v ON_ERROR_STOP=1 -f scripts/normalize_table_owners.sql

DO $$
DECLARE
  target_role text := 'nark';
  r record;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = target_role) THEN
    RAISE EXCEPTION 'role % does not exist', target_role;
  END IF;

  -- Tables first (cascades to owned sequences)
  FOR r IN
    SELECT c.relname
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'r'
      AND pg_get_userbyid(c.relowner) <> target_role
    ORDER BY c.relname
  LOOP
    EXECUTE format('ALTER TABLE public.%I OWNER TO %I', r.relname, target_role);
    RAISE NOTICE 'table public.% → %', r.relname, target_role;
  END LOOP;

  -- Any remaining sequences (not owned by a table, or still wrong owner)
  FOR r IN
    SELECT c.relname
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'S'
      AND pg_get_userbyid(c.relowner) <> target_role
    ORDER BY c.relname
  LOOP
    EXECUTE format('ALTER SEQUENCE public.%I OWNER TO %I', r.relname, target_role);
    RAISE NOTICE 'sequence public.% → %', r.relname, target_role;
  END LOOP;
END $$;

-- Verify: should return 0 rows
SELECT c.relname AS name,
       CASE c.relkind WHEN 'r' THEN 'table' ELSE 'sequence' END AS kind,
       pg_get_userbyid(c.relowner) AS owner
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
  AND c.relkind IN ('r', 'S')
  AND pg_get_userbyid(c.relowner) <> 'nark'
ORDER BY kind, name;

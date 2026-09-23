-- D5: Rename currencies.is_default → is_system (system catalog flag, not user primary).
-- Idempotent. Wallet/category is_default unchanged.
--
--   psql -d pandapocket -U nark -v ON_ERROR_STOP=1 -f scripts/rename_currency_is_system.sql

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'currencies'
      AND column_name = 'is_default'
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'currencies'
      AND column_name = 'is_system'
  ) THEN
    ALTER TABLE public.currencies RENAME COLUMN is_default TO is_system;
    RAISE NOTICE 'currencies.is_default → is_system';
  ELSE
    RAISE NOTICE 'currencies.is_system already present (or is_default missing); skip rename';
  END IF;
END $$;

-- Verify distribution
SELECT is_system, (user_id IS NULL) AS is_catalog, COUNT(*)
FROM currencies
GROUP BY 1, 2
ORDER BY 1, 2;

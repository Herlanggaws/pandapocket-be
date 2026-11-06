-- One-off: mark known internal / throwaway accounts as users.is_internal = true.
-- Pattern-only WHERE clauses (no hardcoded full emails). Postgres.
-- Do NOT run against prod without review. Dry run: replace COMMIT with ROLLBACK.

BEGIN;

-- Preview how many rows would be updated by the patterns below.
SELECT count(*) AS would_mark_internal
FROM users
WHERE deleted_at IS NULL
  AND is_internal = false
  AND (
    -- Owner accounts: local part starts with herlangga (any domain).
    lower(email) LIKE 'herlangga%'
    -- Single-letter local part (a@… style throwaways).
    OR lower(email) ~ '^[a-z]@'
    -- Local part starts with abc (abc…@…).
    OR lower(email) ~ '^abc'
    -- Local part starts with 123 at mail.com.
    OR lower(email) ~ '^123[^@]*@mail\.com$'
    -- Plus-address +test / +test… before @.
    OR lower(email) ~ '\+test[^@]*@'
  );

UPDATE users
SET is_internal = true
WHERE deleted_at IS NULL
  AND is_internal = false
  AND (
    -- Owner accounts: local part starts with herlangga (any domain).
    lower(email) LIKE 'herlangga%'
    -- Single-letter local part (a@… style throwaways).
    OR lower(email) ~ '^[a-z]@'
    -- Local part starts with abc (abc…@…).
    OR lower(email) ~ '^abc'
    -- Local part starts with 123 at mail.com.
    OR lower(email) ~ '^123[^@]*@mail\.com$'
    -- Plus-address +test / +test… before @.
    OR lower(email) ~ '\+test[^@]*@'
  );

COMMIT;

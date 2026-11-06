-- Backfill users.created_at (Go zero time) and users.last_login_at (never written).
-- Also backfill zero created_at on expenses/incomes (GORM Save wiped them on update).
-- Postgres. Local/ops draft only — do NOT run against prod without review.
-- Dry run: replace the final COMMIT with ROLLBACK.

BEGIN;

-- ---------------------------------------------------------------------------
-- (a) created_at: users with Go zero / pre-epoch timestamps
-- Prefer earliest SSO identity, else earliest expense/income row, else updated_at.
-- Table names from models: users, user_identities, expenses, incomes.
-- Only apply when the computed value is finite; infinity means no source existed.
-- Every MIN(...) source ignores timestamps < 1971-01-01 so zeroed child rows
-- cannot win LEAST and re-poison users.created_at.
-- ---------------------------------------------------------------------------

CREATE TEMP TABLE tmp_created_at_candidates ON COMMIT DROP AS
WITH candidates AS (
  SELECT
    u.id,
    u.email,
    u.created_at AS current_created_at,
    u.updated_at,
    (
      SELECT MIN(ui.created_at)
      FROM user_identities ui
      WHERE ui.user_id = u.id
        AND ui.created_at >= TIMESTAMPTZ '1971-01-01'
    ) AS first_identity_at,
    (
      SELECT MIN(t.created_at)
      FROM (
        SELECT e.created_at
        FROM expenses e
        WHERE e.user_id = u.id
          AND e.created_at >= TIMESTAMPTZ '1971-01-01'
        UNION ALL
        SELECT i.created_at
        FROM incomes i
        WHERE i.user_id = u.id
          AND i.created_at >= TIMESTAMPTZ '1971-01-01'
      ) t
    ) AS first_txn_at,
    LEAST(
      COALESCE(
        (
          SELECT MIN(ui.created_at)
          FROM user_identities ui
          WHERE ui.user_id = u.id
            AND ui.created_at >= TIMESTAMPTZ '1971-01-01'
        ),
        'infinity'::timestamptz
      ),
      COALESCE(
        (
          SELECT MIN(t.created_at)
          FROM (
            SELECT e.created_at
            FROM expenses e
            WHERE e.user_id = u.id
              AND e.created_at >= TIMESTAMPTZ '1971-01-01'
            UNION ALL
            SELECT i.created_at
            FROM incomes i
            WHERE i.user_id = u.id
              AND i.created_at >= TIMESTAMPTZ '1971-01-01'
          ) t
        ),
        'infinity'::timestamptz
      ),
      COALESCE(
        CASE
          WHEN u.updated_at >= TIMESTAMPTZ '1971-01-01' THEN u.updated_at
        END,
        'infinity'::timestamptz
      )
    ) AS proposed_created_at
  FROM users u
  WHERE u.created_at < TIMESTAMPTZ '1971-01-01'
)
SELECT * FROM candidates;

-- Preview (resolvable: proposed_created_at is finite)
SELECT
  id,
  email,
  current_created_at,
  updated_at,
  first_identity_at,
  first_txn_at,
  proposed_created_at
FROM tmp_created_at_candidates
WHERE proposed_created_at <> 'infinity'::timestamptz
ORDER BY id;

-- Preview (unresolvable: no identities, no expenses/incomes, NULL/invalid updated_at).
-- These need manual handling — do not auto-set created_at to infinity.
SELECT
  id,
  email,
  current_created_at,
  updated_at,
  proposed_created_at
FROM tmp_created_at_candidates
WHERE proposed_created_at = 'infinity'::timestamptz
ORDER BY id;

-- Apply (finite proposed values only)
UPDATE users u
SET created_at = c.proposed_created_at
FROM tmp_created_at_candidates c
WHERE u.id = c.id
  AND c.proposed_created_at <> 'infinity'::timestamptz;

-- ---------------------------------------------------------------------------
-- (b) last_login_at from session table `tokens` (has created_at).
-- Tokens are inserted on password login, SSO, and /refresh, so this is the
-- best available historical activity signal when last_login_at was never written.
-- ---------------------------------------------------------------------------

-- Preview
SELECT
  u.id,
  u.email,
  u.last_login_at AS current_last_login_at,
  (
    SELECT MAX(t.created_at)
    FROM tokens t
    WHERE t.user_id = u.id
      AND t.created_at >= TIMESTAMPTZ '1971-01-01'
  ) AS proposed_last_login_at
FROM users u
WHERE u.last_login_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM tokens t
    WHERE t.user_id = u.id
      AND t.created_at >= TIMESTAMPTZ '1971-01-01'
  )
ORDER BY u.id;

-- Apply
UPDATE users u
SET last_login_at = (
  SELECT MAX(t.created_at)
  FROM tokens t
  WHERE t.user_id = u.id
    AND t.created_at >= TIMESTAMPTZ '1971-01-01'
)
WHERE u.last_login_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM tokens t
    WHERE t.user_id = u.id
      AND t.created_at >= TIMESTAMPTZ '1971-01-01'
  );

-- ---------------------------------------------------------------------------
-- (c) expenses / incomes with Go zero created_at.
-- Prefer updated_at when valid (>= 1971), else the row's date (date column)
-- cast to timestamptz at UTC midnight.
-- ---------------------------------------------------------------------------

CREATE TEMP TABLE tmp_txn_created_at_backfill ON COMMIT DROP AS
SELECT
  'expenses'::text AS tbl,
  e.id,
  e.user_id,
  e.date,
  e.created_at AS current_created_at,
  e.updated_at,
  CASE
    WHEN e.updated_at >= TIMESTAMPTZ '1971-01-01' THEN e.updated_at
    ELSE ((e.date)::timestamp AT TIME ZONE 'UTC')
  END AS proposed_created_at
FROM expenses e
WHERE e.created_at < TIMESTAMPTZ '1971-01-01'
UNION ALL
SELECT
  'incomes'::text AS tbl,
  i.id,
  i.user_id,
  i.date,
  i.created_at AS current_created_at,
  i.updated_at,
  CASE
    WHEN i.updated_at >= TIMESTAMPTZ '1971-01-01' THEN i.updated_at
    ELSE ((i.date)::timestamp AT TIME ZONE 'UTC')
  END AS proposed_created_at
FROM incomes i
WHERE i.created_at < TIMESTAMPTZ '1971-01-01';

-- Preview counts
SELECT tbl, COUNT(*) AS rows_to_fix
FROM tmp_txn_created_at_backfill
GROUP BY tbl
ORDER BY tbl;

-- Preview sample
SELECT
  tbl,
  id,
  user_id,
  date,
  current_created_at,
  updated_at,
  proposed_created_at
FROM tmp_txn_created_at_backfill
ORDER BY tbl, id
LIMIT 40;

-- Apply expenses
UPDATE expenses e
SET created_at = b.proposed_created_at
FROM tmp_txn_created_at_backfill b
WHERE b.tbl = 'expenses'
  AND e.id = b.id;

-- Apply incomes
UPDATE incomes i
SET created_at = b.proposed_created_at
FROM tmp_txn_created_at_backfill b
WHERE b.tbl = 'incomes'
  AND i.id = b.id;

COMMIT;
-- Dry run: change COMMIT above to ROLLBACK.

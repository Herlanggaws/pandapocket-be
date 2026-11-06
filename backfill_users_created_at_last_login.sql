-- Backfill users.created_at (Go zero time) and users.last_login_at (never written).
-- Postgres. Local/ops draft only — do NOT run against prod without review.
-- Dry run: replace the final COMMIT with ROLLBACK.

BEGIN;

-- ---------------------------------------------------------------------------
-- (a) created_at: users with Go zero / pre-epoch timestamps
-- Prefer earliest SSO identity, else earliest expense/income row, else updated_at.
-- Table names from models: users, user_identities, expenses, incomes.
-- Only apply when the computed value is finite; infinity means no source existed.
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
    ) AS first_identity_at,
    (
      SELECT MIN(t.created_at)
      FROM (
        SELECT created_at FROM expenses e WHERE e.user_id = u.id
        UNION ALL
        SELECT created_at FROM incomes i WHERE i.user_id = u.id
      ) t
    ) AS first_txn_at,
    LEAST(
      COALESCE(
        (SELECT MIN(ui.created_at) FROM user_identities ui WHERE ui.user_id = u.id),
        'infinity'::timestamptz
      ),
      COALESCE(
        (
          SELECT MIN(t.created_at)
          FROM (
            SELECT created_at FROM expenses e WHERE e.user_id = u.id
            UNION ALL
            SELECT created_at FROM incomes i WHERE i.user_id = u.id
          ) t
        ),
        'infinity'::timestamptz
      ),
      COALESCE(u.updated_at, 'infinity'::timestamptz)
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

-- Preview (unresolvable: no identities, no expenses/incomes, NULL updated_at).
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
  ) AS proposed_last_login_at
FROM users u
WHERE u.last_login_at IS NULL
  AND EXISTS (SELECT 1 FROM tokens t WHERE t.user_id = u.id)
ORDER BY u.id;

-- Apply
UPDATE users u
SET last_login_at = (
  SELECT MAX(t.created_at)
  FROM tokens t
  WHERE t.user_id = u.id
)
WHERE u.last_login_at IS NULL
  AND EXISTS (SELECT 1 FROM tokens t WHERE t.user_id = u.id);

COMMIT;

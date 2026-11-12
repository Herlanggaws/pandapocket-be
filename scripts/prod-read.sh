#!/usr/bin/env bash
# Fixed read-only prod metrics over SSH + local psql.
# Counts only (no PII). Never embeds credentials; set PROD_SSH / PROD_DB_* from your env.
# Connection: SSH to the deploy host, then psql as the DB role (see doc/ops/deploy-prod.md).
set -euo pipefail

DEFAULT_SINCE="2026-10-01"

usage() {
  cat <<'EOF'
Usage: prod-read.sh <subcommand> [options]

Read-only prod count queries via SSH → psql. No free-form SQL.

Subcommands:
  users-summary       total / non-internal / activated / weekly-active user counts
  tx-today            expense+income counts and distinct users for today and prior 3 days (Asia/Jakarta)
  google-onboarding   Google-auth users since a date: totals, 0 wallets, 0 txs, onboarding done
  subs-summary        trial/Pro counts, month paid Pro + revenue (WIB), trial starts/conversions
  retention-input     cohort retention weeks 1–4 + all-real-users weekly-active median
  liabilities-usage   liability/receivable adoption among real users (+ weekly-active)

Options:
  --since YYYY-MM-DD  only for google-onboarding (default: 2026-10-01)
  --min-days N        only for retention-input; cohort newer bound (default: 30)
  --max-days N        only for retention-input; cohort older bound (default: 45)
  -h, --help          show this help

Required env (no credentials in this script):
  PROD_SSH       SSH target, e.g. nark@host
  PROD_DB_NAME   database name, e.g. pandapocket
  PROD_DB_USER   database role (prefer the SELECT-only role from
                 scripts/sql/create_readonly_role.sql; migrate role also works)

Optional:
  PROD_DB_HOST   psql -h on the remote (default: omit → peer/unix socket)
  PROD_DB_PORT   psql -p on the remote (default: omit)
EOF
}

require_env() {
  local missing=0
  local var
  for var in PROD_SSH PROD_DB_NAME PROD_DB_USER; do
    if [[ -z "${!var:-}" ]]; then
      echo "error: $var is not set" >&2
      missing=1
    fi
  done
  if [[ "$missing" -ne 0 ]]; then
    echo "Set env vars from your local secret store; this script never reads credential files." >&2
    exit 1
  fi
}

validate_since() {
  local value="$1"
  if [[ ! "$value" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
    echo "error: --since must be YYYY-MM-DD (got: $value)" >&2
    exit 1
  fi
}

validate_positive_int() {
  local name="$1"
  local value="$2"
  if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then
    echo "error: $name must be a positive integer (got: $value)" >&2
    exit 1
  fi
}

remote_psql_cmd() {
  local cmd
  cmd="psql -d $(printf '%q' "$PROD_DB_NAME") -U $(printf '%q' "$PROD_DB_USER") -v ON_ERROR_STOP=1 -P pager=off"
  if [[ -n "${PROD_DB_HOST:-}" ]]; then
    cmd+=" -h $(printf '%q' "$PROD_DB_HOST")"
  fi
  if [[ -n "${PROD_DB_PORT:-}" ]]; then
    cmd+=" -p $(printf '%q' "$PROD_DB_PORT")"
  fi
  printf '%s' "$cmd"
}

# Run SQL on the remote host inside BEGIN READ ONLY … ROLLBACK.
# Arguments after flags are forwarded to remote psql (e.g. -Atc for probes).
# Body SQL is read from stdin when no -c/-f style one-shot is intended via heredoc to ssh.
run_readonly_sql() {
  local extra_args=("$@")
  local remote
  remote="$(remote_psql_cmd)"
  if [[ ${#extra_args[@]} -gt 0 ]]; then
    remote+=" $(printf '%q ' "${extra_args[@]}")"
  fi

  # shellcheck disable=SC2029
  ssh -o BatchMode=yes "$PROD_SSH" "$remote" <<SQL
SET default_transaction_read_only = on;
BEGIN READ ONLY;
$(cat)
ROLLBACK;
SQL
}

column_exists() {
  local table="$1"
  local column="$2"
  local result
  result="$(
    run_readonly_sql -At <<SQL
SELECT EXISTS (
  SELECT 1
  FROM information_schema.columns
  WHERE table_schema = 'public'
    AND table_name = '$(printf '%s' "$table" | sed "s/'/''/g")"
    AND column_name = '$(printf '%s' "$column" | sed "s/'/''/g")"
);
SQL
  )"
  result="$(printf '%s' "$result" | tr -d '[:space:]')"
  [[ "$result" == "t" ]]
}

cmd_users_summary() {
  local non_internal_select
  if column_exists users is_internal; then
    non_internal_select="(
      SELECT COUNT(*)::text
      FROM users u
      WHERE u.deleted_at IS NULL
        AND COALESCE(u.is_internal, false) = false
    )"
  else
    non_internal_select="'n/a'"
  fi

  run_readonly_sql <<SQL
\\pset format aligned
\\pset tuples_only off

WITH active_users AS (
  SELECT u.id, u.created_at
  FROM users u
  WHERE u.deleted_at IS NULL
),
tx AS (
  SELECT e.user_id, e.created_at
  FROM expenses e
  UNION ALL
  SELECT i.user_id, i.created_at
  FROM incomes i
)
SELECT
  (SELECT COUNT(*) FROM active_users) AS total_users,
  ${non_internal_select} AS non_internal_users,
  (
    SELECT COUNT(DISTINCT au.id)
    FROM active_users au
    WHERE EXISTS (
      SELECT 1
      FROM tx
      WHERE tx.user_id = au.id
        AND tx.created_at >= au.created_at
        AND tx.created_at < au.created_at + INTERVAL '7 days'
    )
  ) AS activated_users,
  (
    SELECT COUNT(DISTINCT tx.user_id)
    FROM tx
    WHERE tx.created_at >= (NOW() - INTERVAL '7 days')
      AND EXISTS (
        SELECT 1 FROM active_users au WHERE au.id = tx.user_id
      )
  ) AS weekly_active_users;
SQL
}

cmd_tx_today() {
  run_readonly_sql <<'SQL'
\pset format aligned
\pset tuples_only off

WITH bounds AS (
  SELECT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date AS today_jkt
),
days AS (
  SELECT (b.today_jkt - g.n) AS day_jkt
  FROM bounds b
  CROSS JOIN generate_series(0, 3) AS g(n)
),
tx AS (
  SELECT e.user_id, (e.created_at AT TIME ZONE 'Asia/Jakarta')::date AS day_jkt
  FROM expenses e
  UNION ALL
  SELECT i.user_id, (i.created_at AT TIME ZONE 'Asia/Jakarta')::date AS day_jkt
  FROM incomes i
)
SELECT
  d.day_jkt::text AS day_asia_jakarta,
  COUNT(t.user_id) AS transaction_count,
  COUNT(DISTINCT t.user_id) AS distinct_users
FROM days d
LEFT JOIN tx t ON t.day_jkt = d.day_jkt
GROUP BY d.day_jkt
ORDER BY d.day_jkt DESC;
SQL
}

onboarding_completed_expr() {
  # Emit a boolean SQL expression for "user id = g.id completed onboarding".
  if column_exists users onboarding_completed; then
    cat <<'EXPR'
EXISTS (
  SELECT 1 FROM users u
  WHERE u.id = g.id AND u.onboarding_completed IS TRUE
)
EXPR
    return
  fi
  if column_exists user_preferences onboarding_completed; then
    cat <<'EXPR'
EXISTS (
  SELECT 1 FROM user_preferences up
  WHERE up.user_id = g.id AND up.onboarding_completed IS TRUE
)
EXPR
    return
  fi
  if column_exists user_preferences onboarding; then
    cat <<'EXPR'
EXISTS (
  SELECT 1
  FROM user_preferences up
  WHERE up.user_id = g.id
    AND COALESCE(
      NULLIF(lower(up.onboarding::json ->> 'onboarding_completed'), ''),
      'false'
    ) IN ('true', '1', 't', 'yes')
)
EXPR
    return
  fi
  printf 'false\n'
}

cmd_google_onboarding() {
  local since="$1"
  validate_since "$since"

  local completed_expr
  completed_expr="$(onboarding_completed_expr)"

  run_readonly_sql <<SQL
\\pset format aligned
\\pset tuples_only off

WITH google_users AS (
  SELECT DISTINCT u.id
  FROM users u
  INNER JOIN user_identities ui
    ON ui.user_id = u.id
   AND ui.provider = 'google'
  WHERE u.deleted_at IS NULL
    AND u.created_at >= DATE '${since}'
),
wallet_counts AS (
  SELECT w.user_id, COUNT(*) AS n
  FROM wallets w
  INNER JOIN google_users g ON g.id = w.user_id
  GROUP BY w.user_id
),
tx_counts AS (
  SELECT x.user_id, COUNT(*) AS n
  FROM (
    SELECT e.user_id FROM expenses e
    UNION ALL
    SELECT i.user_id FROM incomes i
  ) x
  INNER JOIN google_users g ON g.id = x.user_id
  GROUP BY x.user_id
),
onboarding AS (
  SELECT
    g.id AS user_id,
    (${completed_expr}) AS completed
  FROM google_users g
)
SELECT
  (SELECT COUNT(*) FROM google_users) AS google_users,
  (
    SELECT COUNT(*) FROM google_users g
    WHERE NOT EXISTS (SELECT 1 FROM wallet_counts w WHERE w.user_id = g.id)
  ) AS with_zero_wallets,
  (
    SELECT COUNT(*) FROM google_users g
    WHERE NOT EXISTS (SELECT 1 FROM tx_counts t WHERE t.user_id = g.id)
  ) AS with_zero_transactions,
  (SELECT COUNT(*) FROM onboarding o WHERE o.completed) AS completed_onboarding;
SQL
}

cmd_subs_summary() {
  # Internal exclusion uses users.is_internal (column on main). Trial rows are
  # created at registration, so trial_started uses subscriptions.created_at
  # for rows with trial_ends_at set. Expired/canceled month bucket is an
  # estimate from updated_at (no dedicated event timestamp).
  run_readonly_sql <<'SQL'
\pset format aligned
\pset tuples_only off

WITH bounds AS (
  SELECT
    date_trunc('month', (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')) AS month_start_jkt,
    (date_trunc('month', (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta'))
      + INTERVAL '1 month') AS month_end_jkt
),
month_utc AS (
  SELECT
    (b.month_start_jkt AT TIME ZONE 'Asia/Jakarta') AS month_start,
    (b.month_end_jkt AT TIME ZONE 'Asia/Jakarta') AS month_end
  FROM bounds b
),
eligible_users AS (
  SELECT u.id
  FROM users u
  WHERE u.deleted_at IS NULL
    AND COALESCE(u.is_internal, false) = false
),
subs AS (
  SELECT s.*
  FROM subscriptions s
  INNER JOIN eligible_users eu ON eu.id = s.user_id
),
paid_pro_now AS (
  SELECT s.billing_interval
  FROM subs s
  WHERE s.billing_interval IS NOT NULL
    AND (
      (
        s.current_period_end IS NOT NULL
        AND s.current_period_end > NOW()
        AND s.status IN ('active', 'canceled')
      )
      OR (
        s.status = 'past_due'
        AND s.grace_ends_at IS NOT NULL
        AND s.grace_ends_at > NOW()
      )
    )
),
successful_payments AS (
  SELECT
    pp.user_id,
    pp.kind,
    pp.amount,
    COALESCE(
      CASE
        WHEN (e.payload::json->'data'->>'paid_at') ~ '^[0-9]{4}-'
          THEN (e.payload::json->'data'->>'paid_at')::timestamptz
        ELSE NULL
      END,
      e.received_at
    ) AS paid_at
  FROM billing_webhook_events e
  INNER JOIN pending_payments pp
    ON pp.payment_id = (e.payload::json->'data'->>'id')
  INNER JOIN eligible_users eu ON eu.id = pp.user_id
  WHERE e.payload::json->>'type' = 'payment.paid'
    AND COALESCE(e.payload::json->'data'->>'return_url', '')
      NOT ILIKE '%stg.berbudget.com%'
),
month_payments AS (
  SELECT sp.*
  FROM successful_payments sp
  CROSS JOIN month_utc m
  WHERE sp.paid_at >= m.month_start
    AND sp.paid_at < m.month_end
),
trials_started_month AS (
  SELECT s.user_id,
    s.billing_interval,
    s.current_period_end
  FROM subs s
  CROSS JOIN month_utc m
  WHERE s.trial_ends_at IS NOT NULL
    AND s.created_at >= m.month_start
    AND s.created_at < m.month_end
)
SELECT
  'users.is_internal=false'::text AS internal_exclusion,
  (
    SELECT COUNT(*)::bigint
    FROM subs s
    WHERE s.status = 'trialing'
      AND s.trial_ends_at IS NOT NULL
      AND s.trial_ends_at > NOW()
  ) AS trial_active_now,
  (
    SELECT COUNT(*)::bigint
    FROM paid_pro_now p
    WHERE p.billing_interval = 'monthly'
  ) AS pro_active_monthly,
  (
    SELECT COUNT(*)::bigint
    FROM paid_pro_now p
    WHERE p.billing_interval = 'semiannual'
  ) AS pro_active_semiannual,
  (
    SELECT COUNT(*)::bigint
    FROM paid_pro_now p
    WHERE p.billing_interval = 'yearly'
  ) AS pro_active_yearly,
  (
    SELECT COUNT(DISTINCT mp.user_id)::bigint
    FROM month_payments mp
    WHERE mp.kind = 'pro'
  ) AS new_paid_pro_this_month,
  (
    SELECT COUNT(*)::bigint
    FROM subs s
    CROSS JOIN month_utc m
    WHERE s.updated_at >= m.month_start
      AND s.updated_at < m.month_end
      AND (
        s.status = 'canceled'
        OR (
          s.status = 'expired'
          AND s.billing_interval IS NOT NULL
        )
      )
  ) AS pro_expired_or_canceled_this_month_est,
  (
    SELECT COALESCE(SUM(mp.amount), 0)::bigint
    FROM month_payments mp
  ) AS total_paid_amount_idr_this_month,
  (
    SELECT COUNT(*)::bigint
    FROM trials_started_month
  ) AS trial_started_this_month,
  (
    SELECT COUNT(*)::bigint
    FROM trials_started_month t
    WHERE t.billing_interval IS NOT NULL
       OR t.current_period_end IS NOT NULL
  ) AS trial_started_converted_to_paid;
SQL
}

cmd_retention_input() {
  local min_days="$1"
  local max_days="$2"
  validate_positive_int "--min-days" "$min_days"
  validate_positive_int "--max-days" "$max_days"
  if [[ "$max_days" -le "$min_days" ]]; then
    echo "error: --max-days must be greater than --min-days" >&2
    exit 1
  fi

  local is_internal_clause="true"
  if column_exists users is_internal; then
    is_internal_clause="COALESCE(u.is_internal, false) = false"
  fi

  run_readonly_sql <<SQL
\\pset format aligned
\\pset tuples_only off

WITH real_users AS (
  SELECT u.id, u.created_at
  FROM users u
  WHERE u.deleted_at IS NULL
    AND ${is_internal_clause}
),
cohort AS (
  SELECT ru.id, ru.created_at
  FROM real_users ru
  WHERE ru.created_at >= NOW() - INTERVAL '${max_days} days'
    AND ru.created_at < NOW() - INTERVAL '${min_days} days'
),
tx AS (
  SELECT e.user_id, e.created_at
  FROM expenses e
  UNION ALL
  SELECT i.user_id, i.created_at
  FROM incomes i
),
cohort_tx AS (
  SELECT
    c.id AS user_id,
    (
      FLOOR(
        EXTRACT(EPOCH FROM (t.created_at - c.created_at))
        / EXTRACT(EPOCH FROM INTERVAL '7 days')
      )::int + 1
    ) AS week_n
  FROM cohort c
  INNER JOIN tx t ON t.user_id = c.id
  WHERE t.created_at >= c.created_at
),
week_user_counts AS (
  SELECT
    ct.week_n,
    ct.user_id,
    COUNT(*)::bigint AS tx_count
  FROM cohort_tx ct
  WHERE ct.week_n BETWEEN 1 AND 4
  GROUP BY ct.week_n, ct.user_id
),
weeks AS (
  SELECT generate_series(1, 4) AS week_n
)
SELECT
  (SELECT COUNT(*)::bigint FROM cohort) AS cohort_users,
  w.week_n,
  COUNT(wuc.user_id)::bigint AS active_users,
  percentile_cont(0.5) WITHIN GROUP (ORDER BY wuc.tx_count::double precision)
    AS median_tx_per_active_user
FROM weeks w
LEFT JOIN week_user_counts wuc ON wuc.week_n = w.week_n
GROUP BY w.week_n
ORDER BY w.week_n;

WITH real_users AS (
  SELECT u.id
  FROM users u
  WHERE u.deleted_at IS NULL
    AND ${is_internal_clause}
),
tx AS (
  SELECT e.user_id, e.created_at
  FROM expenses e
  UNION ALL
  SELECT i.user_id, i.created_at
  FROM incomes i
),
weekly_active AS (
  SELECT t.user_id, COUNT(*)::bigint AS tx_count
  FROM tx t
  INNER JOIN real_users ru ON ru.id = t.user_id
  WHERE t.created_at >= NOW() - INTERVAL '7 days'
  GROUP BY t.user_id
)
SELECT
  COUNT(*)::bigint AS real_users_weekly_active,
  percentile_cont(0.5) WITHIN GROUP (ORDER BY wa.tx_count::double precision)
    AS median_tx_last_7_days
FROM weekly_active wa;
SQL
}

cmd_liabilities_usage() {
  local is_internal_clause="true"
  if column_exists users is_internal; then
    is_internal_clause="COALESCE(u.is_internal, false) = false"
  fi

  run_readonly_sql <<SQL
\\pset format aligned
\\pset tuples_only off

WITH real_users AS (
  SELECT u.id
  FROM users u
  WHERE u.deleted_at IS NULL
    AND ${is_internal_clause}
),
tx AS (
  SELECT e.user_id, e.created_at
  FROM expenses e
  UNION ALL
  SELECT i.user_id, i.created_at
  FROM incomes i
),
weekly_active AS (
  SELECT DISTINCT t.user_id
  FROM tx t
  INNER JOIN real_users ru ON ru.id = t.user_id
  WHERE t.created_at >= NOW() - INTERVAL '7 days'
),
liability_users AS (
  SELECT l.user_id, BOOL_OR(NOT l.is_archived) AS has_active
  FROM liabilities l
  INNER JOIN real_users ru ON ru.id = l.user_id
  GROUP BY l.user_id
),
receivable_users AS (
  SELECT r.user_id, BOOL_OR(NOT r.is_archived) AS has_active
  FROM receivables r
  INNER JOIN real_users ru ON ru.id = r.user_id
  GROUP BY r.user_id
)
SELECT
  (SELECT COUNT(*)::bigint FROM liability_users) AS users_with_liability,
  (
    SELECT COUNT(*)::bigint FROM liability_users lu WHERE lu.has_active
  ) AS users_with_active_liability,
  (
    SELECT COUNT(*)::bigint
    FROM liability_users lu
    WHERE EXISTS (
      SELECT 1 FROM weekly_active wa WHERE wa.user_id = lu.user_id
    )
  ) AS liability_users_weekly_active,
  (SELECT COUNT(*)::bigint FROM receivable_users) AS users_with_receivable,
  (
    SELECT COUNT(*)::bigint FROM receivable_users ru WHERE ru.has_active
  ) AS users_with_active_receivable,
  (
    SELECT COUNT(*)::bigint
    FROM receivable_users ru
    WHERE EXISTS (
      SELECT 1 FROM weekly_active wa WHERE wa.user_id = ru.user_id
    )
  ) AS receivable_users_weekly_active,
  (
    SELECT COUNT(*)::bigint
    FROM liability_users lu
    WHERE EXISTS (
      SELECT 1 FROM receivable_users ru WHERE ru.user_id = lu.user_id
    )
  ) AS users_with_both,
  (SELECT COUNT(*)::bigint FROM weekly_active) AS real_users_weekly_active;
SQL
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  local subcommand="$1"
  shift

  local since="$DEFAULT_SINCE"
  local min_days=30
  local max_days=45

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --since)
        if [[ $# -lt 2 ]]; then
          echo "error: --since requires YYYY-MM-DD" >&2
          exit 1
        fi
        since="$2"
        shift 2
        ;;
      --min-days)
        if [[ $# -lt 2 ]]; then
          echo "error: --min-days requires a positive integer" >&2
          exit 1
        fi
        min_days="$2"
        shift 2
        ;;
      --max-days)
        if [[ $# -lt 2 ]]; then
          echo "error: --max-days requires a positive integer" >&2
          exit 1
        fi
        max_days="$2"
        shift 2
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        echo "error: unknown option: $1" >&2
        usage >&2
        exit 1
        ;;
    esac
  done

  case "$subcommand" in
    -h|--help)
      usage
      exit 0
      ;;
    users-summary)
      require_env
      cmd_users_summary
      ;;
    tx-today)
      require_env
      cmd_tx_today
      ;;
    google-onboarding)
      validate_since "$since"
      require_env
      cmd_google_onboarding "$since"
      ;;
    subs-summary)
      require_env
      cmd_subs_summary
      ;;
    retention-input)
      require_env
      cmd_retention_input "$min_days" "$max_days"
      ;;
    liabilities-usage)
      require_env
      cmd_liabilities_usage
      ;;
    *)
      echo "error: unknown subcommand: $subcommand" >&2
      usage >&2
      exit 1
      ;;
  esac
}

main "$@"

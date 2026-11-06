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

Options:
  --since YYYY-MM-DD  only for google-onboarding (default: 2026-10-01)
  -h, --help          show this help

Required env (no credentials in this script):
  PROD_SSH       SSH target, e.g. nark@host
  PROD_DB_NAME   database name, e.g. pandapocket
  PROD_DB_USER   database role, e.g. nark

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

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  local subcommand="$1"
  shift

  local since="$DEFAULT_SINCE"

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
    *)
      echo "error: unknown subcommand: $subcommand" >&2
      usage >&2
      exit 1
      ;;
  esac
}

main "$@"

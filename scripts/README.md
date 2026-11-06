## Prod read-only metrics (`prod-read.sh`)

Fixed read-only count queries against prod via SSH → `psql` on the deploy host (same connection path as [`doc/ops/deploy-prod.md`](../../doc/ops/deploy-prod.md): SSH in, then psql as the DB role). Counts only — no emails, amounts, or free-form SQL.

```bash
export PROD_SSH='nark@HOST'          # from your local secret store; never commit
export PROD_DB_NAME='pandapocket'
export PROD_DB_USER='nark'

./scripts/prod-read.sh users-summary
./scripts/prod-read.sh tx-today
./scripts/prod-read.sh google-onboarding
./scripts/prod-read.sh google-onboarding --since 2026-10-01
./scripts/prod-read.sh subs-summary
./scripts/prod-read.sh --help
```

Subcommands: `users-summary` (totals / non-internal if `users.is_internal` exists / activated within 7d of signup / weekly active), `tx-today` (today + prior 3 days Asia/Jakarta), `google-onboarding` (Google-auth users since a date: zero wallets, zero txs, onboarding completed), `subs-summary` (one SELECT over `subscriptions` + successful `billing_webhook_events`/`pending_payments`: trial active now; paid Pro active by `billing_interval` monthly/semiannual/yearly; new paid Pro + total paid IDR this calendar month Asia/Jakarta; Pro expired/canceled this month; trial started this month and conversions to paid). Internal exclusion: `users.is_internal = false` when the column exists, else email patterns (`herlangga%`, single-letter local, `^abc`, `^123…@mail.com`, `+test` plus-address) — same patterns as `mark_internal_users.sql` on the internal-users branch (users-summary itself only reports `n/a` when the column is missing). Every query runs with `default_transaction_read_only=on` inside `BEGIN READ ONLY` … `ROLLBACK`. The script never reads credential files and never embeds passwords (SSH key / peer auth on the host).

## Foreign keys (D1)

Idempotent script: [`add_foreign_keys.sql`](add_foreign_keys.sql).

Applied on prod `pandapocket` 2026-09-18 after orphan pre-flight (all 0). Re-run safely on other environments:

```bash
psql -d pandapocket -U nark -v ON_ERROR_STOP=1 -f scripts/add_foreign_keys.sql
```

## Normalize table owners (D4)

Idempotent script: [`normalize_table_owners.sql`](normalize_table_owners.sql).

Transfers `public` tables + sequences to role `nark` (deploy/migrate user). Tables first so owned serial sequences cascade. Applied on prod 2026-09-23 (verify: 0 non-`nark` leftovers).

```bash
# Requires superuser (peer via sudo on host):
sudo -u postgres psql -d pandapocket -v ON_ERROR_STOP=1 -f scripts/normalize_table_owners.sql
```

**Convention:** prod/staging AutoMigrate and one-off SQL scripts use `DB_USER=nark`. Do not create new tables as `postgres`.

## Rename currency `is_system` (D5)

Idempotent script: [`rename_currency_is_system.sql`](rename_currency_is_system.sql).

Renames `currencies.is_default` → `is_system` (system catalog flag; user primary stays on `user_preferences.primary_currency_id`). Wallet/category `is_default` unchanged.

```bash
psql -d pandapocket -U nark -v ON_ERROR_STOP=1 -f scripts/rename_currency_is_system.sql
```

Apply **in the same window as deploying** the API that reads `is_system` (rename breaks old binaries).

## Available Scripts

### 1. `run_add_currencies.sh` (Recommended)
A comprehensive shell script that provides multiple options for adding currencies.

**Usage:**
```bash
# From the backend directory
./scripts/run_add_currencies.sh [go|sql|auto]
```

**Options:**
- `go` - Use Go script (recommended)
- `sql` - Use SQL script directly  
- `auto` - Try Go first, fallback to SQL (default)

**Environment Variables:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=herlangga.wicaksono
export DB_PASSWORD=your_password
export DB_NAME=panda_pocket
```

### 2. `add_currencies.go`
A Go script that connects to the database and adds currencies programmatically.

**Usage:**
```bash
go run scripts/add_currencies.go
```

### 3. `add_currencies.sql`
A SQL script that adds currencies directly to the database.

**Usage:**
```bash
psql -h localhost -U herlangga.wicaksono -d panda_pocket -f scripts/add_currencies.sql
```

## Currencies Added

The scripts add 20 default currencies including:

- **Major Currencies**: USD, EUR, GBP, JPY, AUD, CAD, CHF, CNY
- **European Currencies**: SEK, NOK, DKK, PLN, CZK, HUF
- **Asian Currencies**: KRW, SGD, **IDR** (Indonesian Rupiah)
- **Other Currencies**: RUB, BRL, INR

## Features

- ✅ **Duplicate Prevention**: Checks for existing currencies before adding
- ✅ **Error Handling**: Comprehensive error handling and logging
- ✅ **Verification**: Confirms successful insertion
- ✅ **Flexible**: Multiple execution methods
- ✅ **Safe**: Won't duplicate existing currencies

## Troubleshooting

### Database Connection Issues
```bash
# Test database connection
psql -h localhost -U herlangga.wicaksono -d panda_pocket -c "SELECT 1;"
```

### Permission Issues
```bash
# Make script executable
chmod +x scripts/run_add_currencies.sh
```

### Go Dependencies
```bash
# Install required Go packages
go mod tidy
```

## Server Deployment

For server deployment, you can:

1. **Copy the scripts to your server**
2. **Set environment variables**
3. **Run the appropriate script**

```bash
# Example for server deployment
export DB_HOST=your-server-host
export DB_USER=your-db-user
export DB_PASSWORD=your-db-password
export DB_NAME=panda_pocket

./scripts/run_add_currencies.sh go
```

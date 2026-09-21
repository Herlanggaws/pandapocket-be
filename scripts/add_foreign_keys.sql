-- Add foreign keys for pandapocket (D1).
-- Idempotent: skips constraints that already exist.
-- Prefer RESTRICT/NO ACTION on money graph; CASCADE only on auth tokens wiped with the user.
-- Nullable FKs (wallet_id, currency/category user_id) allow NULL.

DO $$
BEGIN
  -- Helper pattern repeated per constraint

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_wallets_user') THEN
    ALTER TABLE wallets
      ADD CONSTRAINT fk_wallets_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_wallets_currency') THEN
    ALTER TABLE wallets
      ADD CONSTRAINT fk_wallets_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_currencies_user') THEN
    ALTER TABLE currencies
      ADD CONSTRAINT fk_currencies_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_categories_user') THEN
    ALTER TABLE categories
      ADD CONSTRAINT fk_categories_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_expenses_user') THEN
    ALTER TABLE expenses
      ADD CONSTRAINT fk_expenses_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_expenses_wallet') THEN
    ALTER TABLE expenses
      ADD CONSTRAINT fk_expenses_wallet
      FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_expenses_category') THEN
    ALTER TABLE expenses
      ADD CONSTRAINT fk_expenses_category
      FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_expenses_currency') THEN
    ALTER TABLE expenses
      ADD CONSTRAINT fk_expenses_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_incomes_user') THEN
    ALTER TABLE incomes
      ADD CONSTRAINT fk_incomes_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_incomes_wallet') THEN
    ALTER TABLE incomes
      ADD CONSTRAINT fk_incomes_wallet
      FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_incomes_category') THEN
    ALTER TABLE incomes
      ADD CONSTRAINT fk_incomes_category
      FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_incomes_currency') THEN
    ALTER TABLE incomes
      ADD CONSTRAINT fk_incomes_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_budgets_user') THEN
    ALTER TABLE budgets
      ADD CONSTRAINT fk_budgets_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_budgets_category') THEN
    ALTER TABLE budgets
      ADD CONSTRAINT fk_budgets_category
      FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_budgets_currency') THEN
    ALTER TABLE budgets
      ADD CONSTRAINT fk_budgets_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_transfers_user') THEN
    ALTER TABLE transfers
      ADD CONSTRAINT fk_transfers_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_transfers_from_wallet') THEN
    ALTER TABLE transfers
      ADD CONSTRAINT fk_transfers_from_wallet
      FOREIGN KEY (from_wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_transfers_to_wallet') THEN
    ALTER TABLE transfers
      ADD CONSTRAINT fk_transfers_to_wallet
      FOREIGN KEY (to_wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_preferences_user') THEN
    ALTER TABLE user_preferences
      ADD CONSTRAINT fk_user_preferences_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_preferences_currency') THEN
    ALTER TABLE user_preferences
      ADD CONSTRAINT fk_user_preferences_currency
      FOREIGN KEY (primary_currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tokens_user') THEN
    ALTER TABLE tokens
      ADD CONSTRAINT fk_tokens_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_password_reset_tokens_user') THEN
    ALTER TABLE password_reset_tokens
      ADD CONSTRAINT fk_password_reset_tokens_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_account_reset_challenges_user') THEN
    ALTER TABLE account_reset_challenges
      ADD CONSTRAINT fk_account_reset_challenges_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_recurring_user') THEN
    ALTER TABLE recurring_transactions
      ADD CONSTRAINT fk_recurring_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_recurring_wallet') THEN
    ALTER TABLE recurring_transactions
      ADD CONSTRAINT fk_recurring_wallet
      FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_recurring_category') THEN
    ALTER TABLE recurring_transactions
      ADD CONSTRAINT fk_recurring_category
      FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_recurring_currency') THEN
    ALTER TABLE recurring_transactions
      ADD CONSTRAINT fk_recurring_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pending_user') THEN
    ALTER TABLE pending_transactions
      ADD CONSTRAINT fk_pending_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pending_wallet') THEN
    ALTER TABLE pending_transactions
      ADD CONSTRAINT fk_pending_wallet
      FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pending_recurring') THEN
    ALTER TABLE pending_transactions
      ADD CONSTRAINT fk_pending_recurring
      FOREIGN KEY (recurring_transaction_id) REFERENCES recurring_transactions(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pending_category') THEN
    ALTER TABLE pending_transactions
      ADD CONSTRAINT fk_pending_category
      FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pending_currency') THEN
    ALTER TABLE pending_transactions
      ADD CONSTRAINT fk_pending_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_financial_goals_user') THEN
    ALTER TABLE financial_goals
      ADD CONSTRAINT fk_financial_goals_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_financial_goals_currency') THEN
    ALTER TABLE financial_goals
      ADD CONSTRAINT fk_financial_goals_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_financial_goals_wallet') THEN
    ALTER TABLE financial_goals
      ADD CONSTRAINT fk_financial_goals_wallet
      FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_assets_user') THEN
    ALTER TABLE assets
      ADD CONSTRAINT fk_assets_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_assets_currency') THEN
    ALTER TABLE assets
      ADD CONSTRAINT fk_assets_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_liabilities_user') THEN
    ALTER TABLE liabilities
      ADD CONSTRAINT fk_liabilities_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_liabilities_currency') THEN
    ALTER TABLE liabilities
      ADD CONSTRAINT fk_liabilities_currency
      FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_liability_payments_liability') THEN
    ALTER TABLE liability_payments
      ADD CONSTRAINT fk_liability_payments_liability
      FOREIGN KEY (liability_id) REFERENCES liabilities(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_liability_payments_user') THEN
    ALTER TABLE liability_payments
      ADD CONSTRAINT fk_liability_payments_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_liability_payments_expense') THEN
    ALTER TABLE liability_payments
      ADD CONSTRAINT fk_liability_payments_expense
      FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_health_score_snapshots_user') THEN
    ALTER TABLE health_score_snapshots
      ADD CONSTRAINT fk_health_score_snapshots_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_user') THEN
    ALTER TABLE notifications
      ADD CONSTRAINT fk_notifications_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_subscriptions_user') THEN
    ALTER TABLE subscriptions
      ADD CONSTRAINT fk_subscriptions_user
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;
END $$;

# Billing checkout flow audit

Read-only audit of Pro checkout from web “bayar”/upgrade through Doit, return URL, webhook, and subscription update. Also covers renewal vs `pending_payments` and trial-ending reminders.

Sources: backend (this tree), web `pandapocket/` (read-only), `doc/billing/doit-integration.md`, `doc/billing/product-decisions.md`. No secrets.

## 1. Checkout steps (ordered)

1. **Web: user taps bayar / upgrade**
   - Settings billing: `pandapocket/app/settings/billing/page.tsx` → `startCheckout(interval)`.
   - Onboarding plan (when enabled): `pandapocket/app/onboarding/plan/page.tsx` → `startProCheckout`.
   - Client: `BillingService.checkout` → `POST /billing/checkout` with `{ interval }` (`pandapocket/lib/services/billing.ts`).
   - Analytics: `checkout_started` with `plan` + list `price_idr` (product price, not a user transaction amount).

2. **API: create checkout**
   - Route: `POST /billing/checkout` (`backend/internal/application/app.go`).
   - Handler: `BillingHandlers.Checkout` (`backend/internal/interfaces/http/handlers/billing_handlers.go`).
   - Use case: `CreateCheckoutUseCase.Execute` (`backend/internal/application/billing/create_checkout_use_case.go`):
     - Resolves amount via `PriceForInterval`.
     - Creates Doit one-shot payment (`doit.Client.CreatePayment`).
     - Saves `pending_payments` (`KindPro`, payment_id, user_id, interval, amount).
     - Returns `hosted_url`, `payment_id`, `reference`.

3. **Redirect to Doit hosted checkout**
   - Web: `window.location.assign(result.hosted_url)` (billing) or `window.location.href = …` (AI top-up paths). Full document navigation; no popup / `target="_blank"` for Pro checkout.
   - Before leave: `markProCheckoutPending` writes `sessionStorage` key `berbudget:checkoutPending` (≈10 min TTL).

4. **Return URL**
   - Not chosen by the browser. Set on payment create from `DOIT_RETURN_URL` (`backend/internal/infrastructure/doit/client.go`).
   - Docs: production-style `https://bearbudget.id/settings/billing` (`doc/billing/doit-integration.md`).
   - Entitlement is **not** unlocked from the return URL.

5. **Webhook `payment.paid`**
   - `POST /webhooks/doit` → `BillingHandlers.HandleDoitWebhook` → `HandleDoitWebhookUseCase.Execute`.
   - Verifies `PayBridge-Signature`; dedups by event id via `billing_webhook_events` (`Exists` then `TryInsert`).
   - Matches `pending_payments` by payment id (or legacy reference path); validates amount/identity; `activatePro` → `Subscription.ActivatePro` + `subs.Save`.
   - Emits PostHog `subscription_started` (plan / is_renewal / amount_bucket) once per successful paid apply (skipped on deduped retries).

6. **Web after return**
   - Billing page polls `GET /me/subscription` while pending-checkout flag is set; shows waiting / timeout copy if webhook lags.
   - Unlock is server-side only (webhook / DB state).

## 2. Mobile / Instagram in-app browser risks

| Area | Behavior | Where users can get lost |
| --- | --- | --- |
| Leave for payment | Full navigation to Doit (`pay.doit.id`). | In-app browsers may block or mangle third-party redirects; no billing-specific fallback. |
| Return | Lands on `DOIT_RETURN_URL` (settings/billing). Pro unlock waits on webhook. | Paid but delayed/failed webhook → “waiting” UI; user may think payment failed. |
| `sessionStorage` pending flag | Per-tab, same-origin. | New tab, cleared webview storage, or different origin → lose “awaiting payment” banner path; subscription may still update after refresh/poll if webhook succeeded. |
| Auth / cookies | JWT in `localStorage`; fetch uses `credentials: "include"`. | Aggressive in-app storage clearing → logged out on return; must sign in again to see Pro. |
| Instagram / FB / TikTok webview | `InAppBrowserBanner` warns about Google sign-in and can open Chrome intent on Android — used on login/onboarding, **not** on `/settings/billing`. | Checkout started only inside IG webview has no dedicated “open in browser” prompt for Doit. |
| Popups | Not used for Pro checkout. | — |

## 3. Automatic renewals vs `pending_payments`

**No — automatic renewals do not create `pending_payments` rows.**

- Recurring Doit `/v1/billing/subscriptions` is deferred; period end is set from one-shot `payment.paid` (`doc/billing/doit-integration.md`).
- `pending_payments` writers: Pro `CreateCheckoutUseCase` and AI `CreateTopupUseCase` only.
- Hourly `ProcessBillingSubscriptionsUseCase` (`process_billing_subscriptions_use_case.go`, scheduled from `app.go`) only runs `ApplyBillingTransitions` (trial expiry, cancel-at-period-end → free, period end → `past_due` + grace, grace end → free). No Doit charge and no `PendingPaymentRepository.Save`.

Manual re-checkout from settings still creates a new pending row + Doit payment like the first purchase.

## 4. Trial-ending reminder

### Exists (in-app banner only)

| Channel | What | Paths |
| --- | --- | --- |
| Banner | `TrialBanner` — days remaining (calm / urgent / last day), CTA to `/settings/billing`. Shown when `isActiveTrial`. Mounted in `AppShell`. | `pandapocket/components/TrialBanner.tsx`, `pandapocket/components/AppShell.tsx`, helpers `pandapocket/lib/trial.ts`, copy `pandapocket/lib/i18n/messages/id.ts` + `en.ts` |
| Billing UI | Trial end date / days remaining when `status === "trialing"`. | `pandapocket/app/settings/billing/page.tsx` |
| Product intent | Strip messaging aligned with decisions. | `doc/billing/product-decisions.md` (Trial UX messaging) |

### Does not exist

- **Email** trial-ending reminder.
- **Push** trial-ending reminder.
- **Cron / job dedicated to reminders** — only hourly subscription maintenance that **expires** trial state (`ExpireTrialIfNeeded` in `subscription.go` via `process_billing_subscriptions_use_case.go`), with no notification send.

**Verdict:** trial-ending reminder = **in-app banner only** (plus billing settings copy). No email/push/reminder cron.

## 5. Related tracking note (this branch)

Backend emits `subscription_started` on successful Pro `payment.paid` apply with properties `plan` (`monthly` / `6month` / `yearly`), `is_renewal` (best effort: prior paid period or no matching `pending_payments`), and `amount_bucket` only (coarse IDR bands: `lt_50k` / `50k_150k` / `gte_150k` for list prices 19k / 99k / 149k). Deduped webhook retries do not re-capture.

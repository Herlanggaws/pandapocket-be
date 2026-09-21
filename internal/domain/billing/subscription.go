package billing

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("subscription not found")

type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
)

const TrialDurationDays = 14

func (p Plan) String() string { return string(p) }

type BillingInterval string

const (
	IntervalMonthly BillingInterval = "monthly"
	IntervalYearly  BillingInterval = "yearly"
)

func (i BillingInterval) String() string { return string(i) }

type Status string

const (
	StatusTrialing Status = "trialing"
	StatusActive   Status = "active"
	StatusPastDue  Status = "past_due"
	StatusCanceled Status = "canceled"
	StatusExpired  Status = "expired"
)

func (s Status) String() string { return string(s) }

type SubscriptionID struct {
	value int
}

func NewSubscriptionID(id int) SubscriptionID {
	return SubscriptionID{value: id}
}

func (id SubscriptionID) Value() int { return id.value }

type Subscription struct {
	id                 SubscriptionID
	userID             int
	plan               Plan
	billingInterval    *BillingInterval
	status             Status
	trialEndsAt        *time.Time
	currentPeriodEnd   *time.Time
	graceEndsAt        *time.Time
	doitSubscriptionID *string
	doitCustomerRef    string
	cancelAtPeriodEnd  bool
	createdAt          time.Time
	updatedAt          time.Time
}

func CustomerRef(userID int) string {
	return fmt.Sprintf("user:%d", userID)
}

// NewFreeSubscription creates a Free row without trial (backfill / lazy repair).
func NewFreeSubscription(userID int) (*Subscription, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	now := time.Now().UTC()
	return &Subscription{
		userID:            userID,
		plan:              PlanFree,
		status:            StatusExpired,
		doitCustomerRef:   CustomerRef(userID),
		cancelAtPeriodEnd: false,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

// NewTrialSubscription creates a 14-day Pro trial row for a newly registered user.
func NewTrialSubscription(userID int) (*Subscription, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	now := time.Now().UTC()
	trialEnds := now.AddDate(0, 0, TrialDurationDays)
	return &Subscription{
		userID:            userID,
		plan:              PlanFree,
		status:            StatusTrialing,
		trialEndsAt:       &trialEnds,
		doitCustomerRef:   CustomerRef(userID),
		cancelAtPeriodEnd: false,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

func ReconstituteSubscription(
	id SubscriptionID,
	userID int,
	plan Plan,
	billingInterval *BillingInterval,
	status Status,
	trialEndsAt, currentPeriodEnd, graceEndsAt *time.Time,
	doitSubscriptionID *string,
	doitCustomerRef string,
	cancelAtPeriodEnd bool,
	createdAt, updatedAt time.Time,
) *Subscription {
	return &Subscription{
		id:                 id,
		userID:             userID,
		plan:               plan,
		billingInterval:    billingInterval,
		status:             status,
		trialEndsAt:        trialEndsAt,
		currentPeriodEnd:   currentPeriodEnd,
		graceEndsAt:        graceEndsAt,
		doitSubscriptionID: doitSubscriptionID,
		doitCustomerRef:    doitCustomerRef,
		cancelAtPeriodEnd:  cancelAtPeriodEnd,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
	}
}

func (s *Subscription) ID() SubscriptionID                { return s.id }
func (s *Subscription) UserID() int                       { return s.userID }
func (s *Subscription) Plan() Plan                        { return s.plan }
func (s *Subscription) BillingInterval() *BillingInterval { return s.billingInterval }
func (s *Subscription) Status() Status                    { return s.status }
func (s *Subscription) TrialEndsAt() *time.Time           { return s.trialEndsAt }
func (s *Subscription) CurrentPeriodEnd() *time.Time      { return s.currentPeriodEnd }
func (s *Subscription) GraceEndsAt() *time.Time           { return s.graceEndsAt }
func (s *Subscription) DoitSubscriptionID() *string       { return s.doitSubscriptionID }
func (s *Subscription) DoitCustomerRef() string           { return s.doitCustomerRef }
func (s *Subscription) CancelAtPeriodEnd() bool           { return s.cancelAtPeriodEnd }
func (s *Subscription) CreatedAt() time.Time              { return s.createdAt }
func (s *Subscription) UpdatedAt() time.Time              { return s.updatedAt }

func (s *Subscription) AssignID(id SubscriptionID) {
	s.id = id
}

func (s *Subscription) HasUsedTrial() bool {
	return s.trialEndsAt != nil
}

// ExpireTrialIfNeeded sets status to expired when a trial window has ended without paid access.
// Keeps trial_ends_at so the account cannot receive another trial.
func (s *Subscription) ExpireTrialIfNeeded(now time.Time) bool {
	if s.status != StatusTrialing {
		return false
	}
	if s.trialEndsAt == nil || s.trialEndsAt.After(now) {
		return false
	}
	if s.currentPeriodEnd != nil && s.currentPeriodEnd.After(now) {
		return false
	}
	s.status = StatusExpired
	s.updatedAt = now
	return true
}

const (
	monthlyPeriodDays = 30
	yearlyPeriodDays  = 365
	GraceDurationDays = 7
)

// ActivatePro unlocks paid Pro from a verified payment.paid webhook.
// Period length is +30d (monthly) or +365d (yearly) from paidAt.
func (s *Subscription) ActivatePro(interval BillingInterval, paidAt time.Time) error {
	if interval != IntervalMonthly && interval != IntervalYearly {
		return fmt.Errorf("unsupported billing interval: %s", interval)
	}
	paidAt = paidAt.UTC()
	days := monthlyPeriodDays
	if interval == IntervalYearly {
		days = yearlyPeriodDays
	}
	periodEnd := paidAt.AddDate(0, 0, days)

	s.plan = PlanPro
	s.status = StatusActive
	s.billingInterval = &interval
	s.currentPeriodEnd = &periodEnd
	s.graceEndsAt = nil
	s.cancelAtPeriodEnd = false
	s.updatedAt = paidAt
	return nil
}

// ScheduleCancelAtPeriodEnd keeps Pro until current_period_end, then Free.
func (s *Subscription) ScheduleCancelAtPeriodEnd(now time.Time) error {
	if !s.IsPro(now) {
		return errors.New("no active Pro subscription to cancel")
	}
	if s.currentPeriodEnd == nil || !s.currentPeriodEnd.After(now) {
		return errors.New("no active billing period to cancel")
	}
	s.cancelAtPeriodEnd = true
	s.status = StatusCanceled
	s.updatedAt = now.UTC()
	return nil
}

// MarkPastDue starts a 7-day grace window after an unpaid period end.
func (s *Subscription) MarkPastDue(now time.Time) {
	now = now.UTC()
	graceEnd := now.AddDate(0, 0, GraceDurationDays)
	s.status = StatusPastDue
	s.plan = PlanPro
	s.graceEndsAt = &graceEnd
	s.cancelAtPeriodEnd = false
	s.updatedAt = now
}

// DowngradeToFree clears paid entitlement (data retained).
func (s *Subscription) DowngradeToFree(now time.Time) {
	now = now.UTC()
	s.plan = PlanFree
	s.status = StatusExpired
	s.graceEndsAt = nil
	s.cancelAtPeriodEnd = false
	s.updatedAt = now
}

// ApplyBillingTransitions mutates subscription for trial expiry, cancel end, past_due, and grace end.
// Returns true when the row should be saved.
func (s *Subscription) ApplyBillingTransitions(now time.Time) bool {
	now = now.UTC()
	changed := s.ExpireTrialIfNeeded(now)

	if s.cancelAtPeriodEnd && s.currentPeriodEnd != nil && !s.currentPeriodEnd.After(now) {
		s.DowngradeToFree(now)
		return true
	}

	if s.status == StatusActive && s.currentPeriodEnd != nil && !s.currentPeriodEnd.After(now) {
		s.MarkPastDue(now)
		return true
	}

	if s.status == StatusCanceled && s.currentPeriodEnd != nil && !s.currentPeriodEnd.After(now) {
		s.DowngradeToFree(now)
		return true
	}

	if s.status == StatusPastDue && s.graceEndsAt != nil && !s.graceEndsAt.After(now) {
		s.DowngradeToFree(now)
		return true
	}

	return changed
}

// IsPro returns true when the user currently has Pro entitlement.
func (s *Subscription) IsPro(now time.Time) bool {
	if s.trialEndsAt != nil && s.trialEndsAt.After(now) {
		return true
	}
	if s.currentPeriodEnd != nil && s.currentPeriodEnd.After(now) {
		if s.status == StatusActive || s.status == StatusCanceled {
			return true
		}
	}
	if s.status == StatusPastDue && s.graceEndsAt != nil && s.graceEndsAt.After(now) {
		return true
	}
	return false
}

type SubscriptionRepository interface {
	Save(ctx context.Context, sub *Subscription) error
	FindByUserID(ctx context.Context, userID int) (*Subscription, error)
	ListAll(ctx context.Context) ([]*Subscription, error)
}

package ai

import (
	"context"
	"errors"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/domain/entitlement"
)

type CreditService struct {
	credits domainAI.CreditBalanceRepository
	subs    domainBilling.SubscriptionRepository
}

func NewCreditService(credits domainAI.CreditBalanceRepository, subs domainBilling.SubscriptionRepository) *CreditService {
	return &CreditService{credits: credits, subs: subs}
}

type CreditsView struct {
	Available           int  `json:"available"`
	IncludedUnlocked    int  `json:"included_unlocked"`
	IncludedUsed        int  `json:"included_used"`
	PurchasedRemaining  int  `json:"purchased_remaining"`
	IsTrialing          bool `json:"is_trialing"`
}

func (s *CreditService) isTrialing(ctx context.Context, userID int, now time.Time) (bool, error) {
	sub, err := s.subs.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domainBilling.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return sub.Status() == domainBilling.StatusTrialing && sub.TrialEndsAt() != nil && sub.TrialEndsAt().After(now), nil
}

func (s *CreditService) Ensure(ctx context.Context, userID int) (*domainAI.CreditBalance, bool, error) {
	now := time.Now().UTC()
	trialing, err := s.isTrialing(ctx, userID, now)
	if err != nil {
		return nil, false, err
	}

	balance, err := s.credits.FindByUserID(ctx, userID)
	if errors.Is(err, domainAI.ErrBalanceNotFound) {
		balance = domainAI.NewCreditBalance(userID, trialing)
		if err := s.credits.Save(ctx, balance); err != nil {
			return nil, false, err
		}
		return balance, trialing, nil
	}
	if err != nil {
		return nil, false, err
	}

	before := *balance
	domainAI.SyncUnlockForSubscription(balance, trialing)
	if balance.IncludedUnlocked != before.IncludedUnlocked || balance.IncludedGranted != before.IncludedGranted {
		if err := s.credits.Save(ctx, balance); err != nil {
			return nil, false, err
		}
	}
	return balance, trialing, nil
}

func (s *CreditService) View(ctx context.Context, userID int) (*CreditsView, error) {
	balance, trialing, err := s.Ensure(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &CreditsView{
		Available:          balance.Available(),
		IncludedUnlocked:   balance.IncludedUnlocked,
		IncludedUsed:       balance.IncludedUsed,
		PurchasedRemaining: balance.PurchasedRemaining,
		IsTrialing:         trialing,
	}, nil
}

func (s *CreditService) SpendOne(ctx context.Context, userID int) (*CreditsView, error) {
	balance, trialing, err := s.Ensure(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := balance.Spend(); err != nil {
		return nil, err
	}
	if err := s.credits.Save(ctx, balance); err != nil {
		return nil, err
	}
	return &CreditsView{
		Available:          balance.Available(),
		IncludedUnlocked:   balance.IncludedUnlocked,
		IncludedUsed:       balance.IncludedUsed,
		PurchasedRemaining: balance.PurchasedRemaining,
		IsTrialing:         trialing,
	}, nil
}

func (s *CreditService) UnlockFullIncluded(ctx context.Context, userID int) error {
	balance, err := s.credits.FindByUserID(ctx, userID)
	if errors.Is(err, domainAI.ErrBalanceNotFound) {
		balance = domainAI.NewCreditBalance(userID, false)
		balance.UnlockFullIncluded()
		return s.credits.Save(ctx, balance)
	}
	if err != nil {
		return err
	}
	balance.UnlockFullIncluded()
	return s.credits.Save(ctx, balance)
}

func (s *CreditService) ApplyPurchase(ctx context.Context, userID int, pack, doitPaymentID string) error {
	credits, _, err := domainAI.PackCredits(pack)
	if err != nil {
		// Fallback: infer from amount not available; try known packs only
		return err
	}
	return s.credits.RecordPurchase(ctx, userID, pack, doitPaymentID, credits)
}

func RequireProAI(ctx context.Context, checker entitlement.Checker, userID int) error {
	isPro, err := checker.IsPro(ctx, userID)
	if err != nil {
		return err
	}
	if !isPro {
		return entitlement.RequirePro(entitlement.FeatureAIAdvisor)
	}
	return nil
}

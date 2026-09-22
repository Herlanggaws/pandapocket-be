package finance

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReversePaymentRestoresBalance(t *testing.T) {
	l := &Liability{currentBalance: 700}
	if err := l.ReversePayment(300); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.CurrentBalance() != 1000 {
		t.Fatalf("want balance 1000, got %v", l.CurrentBalance())
	}
}

func TestReversePaymentRejectsNonPositive(t *testing.T) {
	l := &Liability{currentBalance: 700}
	if err := l.ReversePayment(0); err == nil {
		t.Fatal("expected error for zero amount")
	}
	if err := l.ReversePayment(-10); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestApplyThenReversePaymentRestoresProgress(t *testing.T) {
	principal := 1000.0
	l := &Liability{currentBalance: 1000, originalPrincipal: &principal}
	if err := l.ApplyPayment(250); err != nil {
		t.Fatalf("ApplyPayment: %v", err)
	}
	pct := l.PayoffProgressPercent()
	if pct == nil || *pct != 25 {
		t.Fatalf("want 25%% progress after payment, got %v", pct)
	}
	if err := l.ReversePayment(250); err != nil {
		t.Fatalf("ReversePayment: %v", err)
	}
	pct = l.PayoffProgressPercent()
	if pct == nil || *pct != 0 {
		t.Fatalf("want 0%% progress after reverse, got %v", pct)
	}
	if l.CurrentBalance() != 1000 {
		t.Fatalf("want balance 1000, got %v", l.CurrentBalance())
	}
}

type memLiabilityRepo struct {
	byID map[int]*Liability
}

func (r *memLiabilityRepo) Save(_ context.Context, liability *Liability) error {
	if r.byID == nil {
		r.byID = map[int]*Liability{}
	}
	r.byID[liability.ID().Value()] = liability
	return nil
}

func (r *memLiabilityRepo) FindByID(_ context.Context, id LiabilityID) (*Liability, error) {
	l, ok := r.byID[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return l, nil
}

func (r *memLiabilityRepo) FindByUserID(context.Context, UserID, bool) ([]*Liability, error) {
	return nil, nil
}

type memPaymentRepo struct {
	byExpense map[int]*LiabilityPayment
	byID      map[int]*LiabilityPayment
}

func (r *memPaymentRepo) Save(_ context.Context, payment *LiabilityPayment) error {
	if r.byID == nil {
		r.byID = map[int]*LiabilityPayment{}
	}
	if r.byExpense == nil {
		r.byExpense = map[int]*LiabilityPayment{}
	}
	r.byID[payment.ID().Value()] = payment
	if payment.ExpenseID() != nil {
		r.byExpense[*payment.ExpenseID()] = payment
	}
	return nil
}

func (r *memPaymentRepo) FindByLiabilityID(context.Context, LiabilityID) ([]*LiabilityPayment, error) {
	return nil, nil
}

func (r *memPaymentRepo) FindByExpenseID(_ context.Context, expenseID int) (*LiabilityPayment, error) {
	return r.byExpense[expenseID], nil
}

func (r *memPaymentRepo) Delete(_ context.Context, id LiabilityPaymentID) error {
	payment, ok := r.byID[id.Value()]
	if !ok {
		return nil
	}
	delete(r.byID, id.Value())
	if payment.ExpenseID() != nil {
		delete(r.byExpense, *payment.ExpenseID())
	}
	return nil
}

func TestReversePaymentByExpenseIDRestoresBalanceAndDeletesPayment(t *testing.T) {
	principal := 1000.0
	userID := NewUserID(1)
	liabilityID := NewLiabilityID(10)
	expenseID := 42

	liability := ReconstituteLiability(
		liabilityID,
		userID,
		"Loan",
		LiabilityTypeLoan,
		NewCurrencyID(1),
		750,
		"",
		false,
		nil,
		time.Now(),
		LiabilityDebtDetails{OriginalPrincipal: &principal, MinimumPayment: 100},
	)

	payment := ReconstituteLiabilityPayment(
		NewLiabilityPaymentID(7),
		liabilityID,
		userID,
		250,
		time.Now(),
		&expenseID,
		"",
		time.Now(),
	)

	liabilityRepo := &memLiabilityRepo{byID: map[int]*Liability{10: liability}}
	paymentRepo := &memPaymentRepo{
		byID:      map[int]*LiabilityPayment{7: payment},
		byExpense: map[int]*LiabilityPayment{42: payment},
	}
	svc := NewLiabilityService(liabilityRepo, paymentRepo)

	if err := svc.ReversePaymentByExpenseID(context.Background(), userID, expenseID); err != nil {
		t.Fatalf("ReversePaymentByExpenseID: %v", err)
	}

	if liability.CurrentBalance() != 1000 {
		t.Fatalf("want balance 1000, got %v", liability.CurrentBalance())
	}
	if paymentRepo.byExpense[expenseID] != nil {
		t.Fatal("expected payment row deleted")
	}
	if paymentRepo.byID[7] != nil {
		t.Fatal("expected payment id deleted")
	}
}

func TestReversePaymentByExpenseIDNoopWhenNoPayment(t *testing.T) {
	svc := NewLiabilityService(&memLiabilityRepo{}, &memPaymentRepo{})
	if err := svc.ReversePaymentByExpenseID(context.Background(), NewUserID(1), 999); err != nil {
		t.Fatalf("expected no-op success, got %v", err)
	}
}

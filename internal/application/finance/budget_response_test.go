package finance

import (
	"panda-pocket/internal/domain/finance"
	"testing"
	"time"
)

func TestSumExpensesForBudgetFiltersCategoryTypeAndCurrency(t *testing.T) {
	amount, _ := finance.NewMoney(500, finance.NewCurrencyID(1))
	budget, err := finance.ReconstituteBudget(
		finance.NewBudgetID(1),
		finance.NewUserID(1),
		finance.NewCategoryID(10),
		amount,
		finance.BudgetLimitFixed,
		nil,
		finance.BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("budget error: %v", err)
	}

	matchingMoney, _ := finance.NewMoney(40, finance.NewCurrencyID(1))
	otherCurrencyMoney, _ := finance.NewMoney(25, finance.NewCurrencyID(2))
	otherCategoryMoney, _ := finance.NewMoney(15, finance.NewCurrencyID(1))
	incomeMoney, _ := finance.NewMoney(100, finance.NewCurrencyID(1))

	transactions := []*finance.Transaction{
		finance.NewTransaction(
			finance.NewTransactionID(1),
			finance.NewUserID(1),
			finance.NewWalletID(1),
			finance.NewCategoryID(10),
			finance.NewCurrencyID(1),
			matchingMoney,
			"lunch",
			time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			finance.TransactionTypeExpense,
		),
		finance.NewTransaction(
			finance.NewTransactionID(2),
			finance.NewUserID(1),
			finance.NewWalletID(1),
			finance.NewCategoryID(10),
			finance.NewCurrencyID(2),
			otherCurrencyMoney,
			"usd lunch",
			time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
			finance.TransactionTypeExpense,
		),
		finance.NewTransaction(
			finance.NewTransactionID(3),
			finance.NewUserID(1),
			finance.NewWalletID(1),
			finance.NewCategoryID(11),
			finance.NewCurrencyID(1),
			otherCategoryMoney,
			"transport",
			time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
			finance.TransactionTypeExpense,
		),
		finance.NewTransaction(
			finance.NewTransactionID(4),
			finance.NewUserID(1),
			finance.NewWalletID(1),
			finance.NewCategoryID(10),
			finance.NewCurrencyID(1),
			incomeMoney,
			"refund",
			time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			finance.TransactionTypeIncome,
		),
	}

	total := sumExpensesForBudget(transactions, budget)
	if total != 40 {
		t.Fatalf("expected total spent 40, got %v", total)
	}
}

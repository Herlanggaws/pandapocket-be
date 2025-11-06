package finance

import (
	"context"
	"time"
)

// TransactionRepository defines the contract for transaction persistence
type TransactionRepository interface {
	Save(ctx context.Context, transaction *Transaction) error
	FindByID(ctx context.Context, id TransactionID) (*Transaction, error)
	FindByIDAndType(ctx context.Context, id TransactionID, transactionType TransactionType) (*Transaction, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Transaction, error)
	FindByUserIDAndDateRange(ctx context.Context, userID UserID, startDate, endDate time.Time) ([]*Transaction, error)
	FindByUserIDAndCategory(ctx context.Context, userID UserID, categoryID CategoryID) ([]*Transaction, error)
	FindByUserIDWithFilters(ctx context.Context, userID UserID, filters TransactionFilters) ([]*Transaction, int64, error)
	Delete(ctx context.Context, id TransactionID) error
	// Dashboard stats methods
	GetTotalCount(ctx context.Context) (int, error)
	GetTotalExpenses(ctx context.Context) (float64, error)
	GetTotalIncome(ctx context.Context) (float64, error)
}

// CategoryRepository defines the contract for category persistence
type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id CategoryID) (*Category, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Category, error)
	FindByUserIDAndType(ctx context.Context, userID UserID, categoryType CategoryType) ([]*Category, error)
	FindDefaultCategories(ctx context.Context) ([]*Category, error)
	Delete(ctx context.Context, id CategoryID) error
	ExistsByID(ctx context.Context, id CategoryID) (bool, error)
}

// CurrencyRepository defines the contract for currency persistence
type CurrencyRepository interface {
	Save(ctx context.Context, currency *Currency) error
	FindByID(ctx context.Context, id CurrencyID) (*Currency, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Currency, error)
	FindDefaultCurrencies(ctx context.Context) ([]*Currency, error)
	Delete(ctx context.Context, id CurrencyID) error
	ExistsByID(ctx context.Context, id CurrencyID) (bool, error)
	ExistsByCodeAndUserID(ctx context.Context, code string, userID UserID) (bool, error)
	SetUserDefaultCurrency(ctx context.Context, userID UserID, currencyID CurrencyID) error
	GetUserDefaultCurrency(ctx context.Context, userID UserID) (*Currency, error)
}

// BudgetRepository defines the contract for budget persistence
type BudgetRepository interface {
	Save(ctx context.Context, budget *Budget) error
	FindByID(ctx context.Context, id BudgetID) (*Budget, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Budget, error)
	FindByUserIDAndCategory(ctx context.Context, userID UserID, categoryID CategoryID) ([]*Budget, error)
	FindActiveByUserID(ctx context.Context, userID UserID) ([]*Budget, error)
	Delete(ctx context.Context, id BudgetID) error
	// Dashboard stats methods
	GetTotalCount(ctx context.Context) (int, error)
	GetCountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
}

// RecurringTransactionRepository defines the contract for recurring transaction persistence
type RecurringTransactionRepository interface {
	Save(ctx context.Context, recurringTransaction *RecurringTransaction) error
	FindByID(ctx context.Context, id RecurringTransactionID) (*RecurringTransaction, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*RecurringTransaction, error)
	FindActiveByUserID(ctx context.Context, userID UserID) ([]*RecurringTransaction, error)
	FindDueTransactions(ctx context.Context) ([]*RecurringTransaction, error)
	Delete(ctx context.Context, id RecurringTransactionID) error
}

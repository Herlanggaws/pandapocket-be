package database

import (
	"context"
	"log"
	"panda-pocket/internal/domain/finance"
	"sort"
	"time"

	"gorm.io/gorm"
)

// GormTransactionRepository implements the TransactionRepository interface using GORM
type GormTransactionRepository struct {
	db *gorm.DB
}

// NewGormTransactionRepository creates a new GORM transaction repository
func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

// Save saves a transaction to the database
func (r *GormTransactionRepository) Save(ctx context.Context, transaction *finance.Transaction) error {
	// Convert domain transaction to GORM model
	var transactionModel interface{}

	if transaction.Type() == finance.TransactionTypeExpense {
		expenseModel := &Expense{
			UserID:      uint(transaction.UserID().Value()),
			CategoryID:  uint(transaction.CategoryID().Value()),
			CurrencyID:  uint(transaction.CurrencyID().Value()),
			Amount:      transaction.Amount().Amount(),
			Description: transaction.Description(),
			Date:        transaction.Date(),
		}

		if transaction.ID().Value() != 0 {
			expenseModel.ID = uint(transaction.ID().Value())
		}

		transactionModel = expenseModel
	} else {
		incomeModel := &Income{
			UserID:      uint(transaction.UserID().Value()),
			CategoryID:  uint(transaction.CategoryID().Value()),
			CurrencyID:  uint(transaction.CurrencyID().Value()),
			Amount:      transaction.Amount().Amount(),
			Description: transaction.Description(),
			Date:        transaction.Date(),
		}

		if transaction.ID().Value() != 0 {
			incomeModel.ID = uint(transaction.ID().Value())
		}

		transactionModel = incomeModel
	}

	// Save using GORM
	if err := r.db.WithContext(ctx).Save(transactionModel).Error; err != nil {
		return err
	}

	return nil
}

// FindByID finds a transaction by ID
func (r *GormTransactionRepository) FindByID(ctx context.Context, id finance.TransactionID) (*finance.Transaction, error) {
	// Try to find in expenses first
	var expenseModel Expense
	err := r.db.WithContext(ctx).First(&expenseModel, id.Value()).Error
	if err == nil {
		return r.expenseToTransaction(&expenseModel), nil
	}

	// If not found in expenses, try incomes
	var incomeModel Income
	err = r.db.WithContext(ctx).First(&incomeModel, id.Value()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	return r.incomeToTransaction(&incomeModel), nil
}

// FindByUserID finds all transactions for a user
func (r *GormTransactionRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Transaction, error) {
	var transactions []*finance.Transaction

	// Get expenses
	var expenseModels []Expense
	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&expenseModels).Error
	if err != nil {
		return nil, err
	}

	for _, model := range expenseModels {
		transactions = append(transactions, r.expenseToTransaction(&model))
	}

	// Get incomes
	var incomeModels []Income
	err = r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&incomeModels).Error
	if err != nil {
		return nil, err
	}

	for _, model := range incomeModels {
		transactions = append(transactions, r.incomeToTransaction(&model))
	}

	return transactions, nil
}

// FindByUserIDAndDateRange finds transactions for a user within a date range
func (r *GormTransactionRepository) FindByUserIDAndDateRange(ctx context.Context, userID finance.UserID, startDate, endDate time.Time) ([]*finance.Transaction, error) {
	var transactions []*finance.Transaction

	// Get expenses
	var expenseModels []Expense
	err := r.db.WithContext(ctx).Where("user_id = ? AND date BETWEEN ? AND ?", userID.Value(), startDate, endDate).Find(&expenseModels).Error
	if err != nil {
		return nil, err
	}

	for _, model := range expenseModels {
		transactions = append(transactions, r.expenseToTransaction(&model))
	}

	// Get incomes
	var incomeModels []Income
	err = r.db.WithContext(ctx).Where("user_id = ? AND date BETWEEN ? AND ?", userID.Value(), startDate, endDate).Find(&incomeModels).Error
	if err != nil {
		return nil, err
	}

	for _, model := range incomeModels {
		transactions = append(transactions, r.incomeToTransaction(&model))
	}

	return transactions, nil
}

// Delete deletes a transaction by ID
func (r *GormTransactionRepository) Delete(ctx context.Context, id finance.TransactionID) error {
	// Try to delete from expenses first
	result := r.db.WithContext(ctx).Delete(&Expense{}, id.Value())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	// If not found in expenses, try incomes
	result = r.db.WithContext(ctx).Delete(&Income{}, id.Value())
	return result.Error
}

// ExistsByID checks if a transaction exists with the given ID
func (r *GormTransactionRepository) ExistsByID(ctx context.Context, id finance.TransactionID) (bool, error) {
	var count int64

	// Check expenses
	err := r.db.WithContext(ctx).Model(&Expense{}).Where("id = ?", id.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check incomes
	err = r.db.WithContext(ctx).Model(&Income{}).Where("id = ?", id.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindByUserIDAndCategory finds transactions by user ID and category
func (r *GormTransactionRepository) FindByUserIDAndCategory(ctx context.Context, userID finance.UserID, categoryID finance.CategoryID) ([]*finance.Transaction, error) {
	var transactions []*finance.Transaction

	// Get expenses
	var expenseModels []Expense
	err := r.db.WithContext(ctx).Where("user_id = ? AND category_id = ?", userID.Value(), categoryID.Value()).Find(&expenseModels).Error
	if err != nil {
		return nil, err
	}

	for _, model := range expenseModels {
		transactions = append(transactions, r.expenseToTransaction(&model))
	}

	// Get incomes
	var incomeModels []Income
	err = r.db.WithContext(ctx).Where("user_id = ? AND category_id = ?", userID.Value(), categoryID.Value()).Find(&incomeModels).Error
	if err != nil {
		return nil, err
	}

	for _, model := range incomeModels {
		transactions = append(transactions, r.incomeToTransaction(&model))
	}

	return transactions, nil
}

// FindByUserIDWithFilters finds transactions for a user with filters
func (r *GormTransactionRepository) FindByUserIDWithFilters(ctx context.Context, userID finance.UserID, filters finance.TransactionFilters) ([]*finance.Transaction, int64, error) {
	var allTransactions []*finance.Transaction
	var totalCount int64

	// Build base query conditions
	baseConditions := "user_id = ?"
	args := []interface{}{userID.Value()}

	// Apply date range filter
	if filters.StartDate != nil && filters.EndDate != nil {
		baseConditions += " AND date BETWEEN ? AND ?"
		args = append(args, *filters.StartDate, *filters.EndDate)
	} else if filters.StartDate != nil {
		baseConditions += " AND date >= ?"
		args = append(args, *filters.StartDate)
	} else if filters.EndDate != nil {
		baseConditions += " AND date <= ?"
		args = append(args, *filters.EndDate)
	}

	// Apply category filter
	if len(filters.CategoryIDs) > 0 {
		categoryIDs := make([]uint, len(filters.CategoryIDs))
		for i, categoryID := range filters.CategoryIDs {
			categoryIDs[i] = uint(categoryID.Value())
		}
		baseConditions += " AND category_id IN ?"
		args = append(args, categoryIDs)
	}

	// Count total records first
	if filters.TransactionType != nil && *filters.TransactionType == finance.TransactionTypeExpense {
		// Count expenses
		err := r.db.WithContext(ctx).Model(&Expense{}).Where(baseConditions, args...).Count(&totalCount).Error
		if err != nil {
			return nil, 0, err
		}
	} else if filters.TransactionType != nil && *filters.TransactionType == finance.TransactionTypeIncome {
		// Count incomes
		err := r.db.WithContext(ctx).Model(&Income{}).Where(baseConditions, args...).Count(&totalCount).Error
		if err != nil {
			return nil, 0, err
		}
	} else {
		// Count both expenses and incomes
		var expenseCount, incomeCount int64

		err := r.db.WithContext(ctx).Model(&Expense{}).Where(baseConditions, args...).Count(&expenseCount).Error
		if err != nil {
			return nil, 0, err
		}

		err = r.db.WithContext(ctx).Model(&Income{}).Where(baseConditions, args...).Count(&incomeCount).Error
		if err != nil {
			return nil, 0, err
		}

		totalCount = expenseCount + incomeCount
	}

	// Apply pagination and fetch data
	if filters.TransactionType != nil && *filters.TransactionType == finance.TransactionTypeExpense {
		// Query expenses only
		var expenseModels []Expense
		query := r.db.WithContext(ctx).Where(baseConditions, args...).Order("date DESC, created_at DESC")

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}

		err := query.Find(&expenseModels).Error
		if err != nil {
			return nil, 0, err
		}

		for _, model := range expenseModels {
			allTransactions = append(allTransactions, r.expenseToTransaction(&model))
		}
	} else if filters.TransactionType != nil && *filters.TransactionType == finance.TransactionTypeIncome {
		// Query incomes only
		var incomeModels []Income
		query := r.db.WithContext(ctx).Where(baseConditions, args...).Order("date DESC, created_at DESC")

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}

		err := query.Find(&incomeModels).Error
		if err != nil {
			return nil, 0, err
		}

		for _, model := range incomeModels {
			allTransactions = append(allTransactions, r.incomeToTransaction(&model))
		}
	} else {
		// Query both tables and combine results
		// For pagination across two tables, we need a more complex approach
		// We'll fetch from both tables and merge, then apply pagination in memory
		// This is not ideal for very large datasets, but works for most use cases

		var expenseModels []Expense
		var incomeModels []Income

		// Fetch all matching records from both tables
		err := r.db.WithContext(ctx).Where(baseConditions, args...).Order("date DESC, created_at DESC").Find(&expenseModels).Error
		if err != nil {
			return nil, 0, err
		}

		err = r.db.WithContext(ctx).Where(baseConditions, args...).Order("date DESC, created_at DESC").Find(&incomeModels).Error
		if err != nil {
			return nil, 0, err
		}

		// Convert to domain objects
		for _, model := range expenseModels {
			allTransactions = append(allTransactions, r.expenseToTransaction(&model))
		}
		for _, model := range incomeModels {
			allTransactions = append(allTransactions, r.incomeToTransaction(&model))
		}

		// Sort combined results by date DESC, then by created_at DESC
		sort.Slice(allTransactions, func(i, j int) bool {
			if allTransactions[i].Date().Equal(allTransactions[j].Date()) {
				return allTransactions[i].CreatedAt().After(allTransactions[j].CreatedAt())
			}
			return allTransactions[i].Date().After(allTransactions[j].Date())
		})

		// Apply pagination in memory
		start := filters.Offset
		end := start + filters.Limit

		if start >= len(allTransactions) {
			allTransactions = []*finance.Transaction{}
		} else {
			if end > len(allTransactions) {
				end = len(allTransactions)
			}
			allTransactions = allTransactions[start:end]
		}
	}

	return allTransactions, totalCount, nil
}

// Helper methods to convert GORM models to domain transactions
func (r *GormTransactionRepository) expenseToTransaction(expense *Expense) *finance.Transaction {
	transactionID := finance.NewTransactionID(int(expense.ID))
	userID := finance.NewUserID(int(expense.UserID))
	categoryID := finance.NewCategoryID(int(expense.CategoryID))
	currencyID := finance.NewCurrencyID(int(expense.CurrencyID))
	amount, err := finance.NewMoney(expense.Amount, currencyID)
	if err != nil {
		// Log error but continue - this shouldn't happen with valid data
		log.Printf("Warning: failed to create money from expense amount: %v", err)
		amount, _ = finance.NewMoney(0, currencyID) // Fallback to 0 amount
	}

	transaction := finance.NewTransaction(
		transactionID,
		userID,
		categoryID,
		currencyID,
		amount,
		expense.Description,
		expense.Date,
		finance.TransactionTypeExpense,
	)
	return transaction
}

func (r *GormTransactionRepository) incomeToTransaction(income *Income) *finance.Transaction {
	transactionID := finance.NewTransactionID(int(income.ID))
	userID := finance.NewUserID(int(income.UserID))
	categoryID := finance.NewCategoryID(int(income.CategoryID))
	currencyID := finance.NewCurrencyID(int(income.CurrencyID))
	amount, err := finance.NewMoney(income.Amount, currencyID)
	if err != nil {
		// Log error but continue - this shouldn't happen with valid data
		log.Printf("Warning: failed to create money from income amount: %v", err)
		amount, _ = finance.NewMoney(0, currencyID) // Fallback to 0 amount
	}

	transaction := finance.NewTransaction(
		transactionID,
		userID,
		categoryID,
		currencyID,
		amount,
		income.Description,
		income.Date,
		finance.TransactionTypeIncome,
	)
	return transaction
}

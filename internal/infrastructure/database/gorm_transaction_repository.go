package database

import (
	"context"
	"log"
	"panda-pocket/internal/domain/finance"
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
	db := dbFromContext(ctx, r.db)
	// Convert domain transaction to GORM model
	var transactionModel interface{}

	var walletID *uint
	if transaction.WalletID() != nil {
		wid := uint(transaction.WalletID().Value())
		walletID = &wid
	}

	if transaction.Type() == finance.TransactionTypeExpense {
		expenseModel := &Expense{
			UserID:      uint(transaction.UserID().Value()),
			CategoryID:  uint(transaction.CategoryID().Value()),
			CurrencyID:  uint(transaction.CurrencyID().Value()),
			WalletID:    walletID,
			Amount:      transaction.Amount().Amount(),
			IsApproved:  transaction.IsApproved(),
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
			WalletID:    walletID,
			Amount:      transaction.Amount().Amount(),
			IsApproved:  transaction.IsApproved(),
			Description: transaction.Description(),
			Date:        transaction.Date(),
		}

		if transaction.ID().Value() != 0 {
			incomeModel.ID = uint(transaction.ID().Value())
		}

		transactionModel = incomeModel
	}

	// Save using GORM
	if err := db.WithContext(ctx).Save(transactionModel).Error; err != nil {
		return err
	}

	return nil
}

// FindByID finds a transaction by ID (checks expenses first, then incomes)
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

// FindByIDAndType finds a transaction by ID, checking the correct table first based on transaction type
func (r *GormTransactionRepository) FindByIDAndType(ctx context.Context, id finance.TransactionID, transactionType finance.TransactionType) (*finance.Transaction, error) {
	if transactionType == finance.TransactionTypeExpense {
		// Check expenses table first
		var expenseModel Expense
		err := r.db.WithContext(ctx).First(&expenseModel, id.Value()).Error
		if err != nil {
			return nil, err
		}
		return r.expenseToTransaction(&expenseModel), nil
	} else {
		// Check incomes table first
		var incomeModel Income
		err := r.db.WithContext(ctx).First(&incomeModel, id.Value()).Error
		if err != nil {
			return nil, err
		}
		return r.incomeToTransaction(&incomeModel), nil
	}
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

// TransactionView represents a unified view of income and expense transactions
type TransactionView struct {
	ID          uint
	UserID      uint
	CategoryID  uint
	CurrencyID  uint
	WalletID    *uint
	Amount      float64
	IsApproved  bool
	Description string
	Date        time.Time
	CreatedAt   time.Time
	Type        string
}

// FindByUserIDWithFilters finds transactions for a user with filters
func (r *GormTransactionRepository) FindByUserIDWithFilters(ctx context.Context, userID finance.UserID, filters finance.TransactionFilters) ([]*finance.Transaction, int64, error) {
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

	// Apply approval filter
	if filters.IsApproved != nil {
		baseConditions += " AND is_approved = ?"
		args = append(args, *filters.IsApproved)
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

	// Case 1: Only Expense or Only Income requested - use standard GORM
	if filters.TransactionType != nil {
		if *filters.TransactionType == finance.TransactionTypeExpense {
			// Count
			err := r.db.WithContext(ctx).Model(&Expense{}).Where(baseConditions, args...).Count(&totalCount).Error
			if err != nil {
				return nil, 0, err
			}

			// Query
			var expenseModels []Expense
			query := r.db.WithContext(ctx).Where(baseConditions, args...).Order("date DESC, created_at DESC")
			if filters.Limit > 0 {
				query = query.Limit(filters.Limit)
			}
			if filters.Offset > 0 {
				query = query.Offset(filters.Offset)
			}

			err = query.Find(&expenseModels).Error
			if err != nil {
				return nil, 0, err
			}

			var transactions []*finance.Transaction
			for _, model := range expenseModels {
				transactions = append(transactions, r.expenseToTransaction(&model))
			}
			return transactions, totalCount, nil

		} else if *filters.TransactionType == finance.TransactionTypeIncome {
			// Count
			err := r.db.WithContext(ctx).Model(&Income{}).Where(baseConditions, args...).Count(&totalCount).Error
			if err != nil {
				return nil, 0, err
			}

			// Query
			var incomeModels []Income
			query := r.db.WithContext(ctx).Where(baseConditions, args...).Order("date DESC, created_at DESC")
			if filters.Limit > 0 {
				query = query.Limit(filters.Limit)
			}
			if filters.Offset > 0 {
				query = query.Offset(filters.Offset)
			}

			err = query.Find(&incomeModels).Error
			if err != nil {
				return nil, 0, err
			}

			var transactions []*finance.Transaction
			for _, model := range incomeModels {
				transactions = append(transactions, r.incomeToTransaction(&model))
			}
			return transactions, totalCount, nil
		}
	}

	// Case 2: Both requested - use UNION
	// 1. Get Total Count
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

	// 2. Build UNION Query for data
	// We use GORM's subquery capability by passing *gorm.DB objects to the Raw query
	// This automatically handles placeholder (?) parameters and avoids manual string concatenation issues

	expenseQuery := r.db.Model(&Expense{}).
		Select("id, user_id, category_id, currency_id, wallet_id, amount, is_approved, description, date, created_at, 'expense' as type").
		Where(baseConditions, args...)

	incomeQuery := r.db.Model(&Income{}).
		Select("id, user_id, category_id, currency_id, wallet_id, amount, is_approved, description, date, created_at, 'income' as type").
		Where(baseConditions, args...)

	// Combine SQL
	// (? UNION ALL ?) will rely on GORM to expand the queries correctly
	fullSQL := "(?) UNION ALL (?) ORDER BY date DESC, created_at DESC"
	params := []interface{}{expenseQuery, incomeQuery}

	// Apply Limit/Offset to the outer query
	if filters.Limit > 0 {
		fullSQL += " LIMIT ?"
		params = append(params, filters.Limit)
	}
	if filters.Offset > 0 {
		fullSQL += " OFFSET ?"
		params = append(params, filters.Offset)
	}

	var views []TransactionView
	err = r.db.WithContext(ctx).Raw(fullSQL, params...).Scan(&views).Error
	if err != nil {
		return nil, 0, err
	}

	// Convert Views to Domain Transactions
	var transactions []*finance.Transaction
	for _, view := range views {
		tID := finance.NewTransactionID(int(view.ID))
		uID := finance.NewUserID(int(view.UserID))
		cID := finance.NewCategoryID(int(view.CategoryID))
		currID := finance.NewCurrencyID(int(view.CurrencyID))
		amount, _ := finance.NewMoney(view.Amount, currID)

		var tType finance.TransactionType
		if view.Type == "expense" {
			tType = finance.TransactionTypeExpense
		} else {
			tType = finance.TransactionTypeIncome
		}

		var walletID *finance.WalletID
		if view.WalletID != nil {
			wid := finance.NewWalletID(int(*view.WalletID))
			walletID = &wid
		}

		transaction := finance.NewTransaction(
			tID,
			uID,
			cID,
			currID,
			amount,
			view.IsApproved,
			view.Description,
			view.Date,
			tType,
			walletID,
		)
		transactions = append(transactions, transaction)
	}

	return transactions, totalCount, nil
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

	var walletID *finance.WalletID
	if expense.WalletID != nil {
		wid := finance.NewWalletID(int(*expense.WalletID))
		walletID = &wid
	}

	transaction := finance.NewTransaction(
		transactionID,
		userID,
		categoryID,
		currencyID,
		amount,
		expense.IsApproved,
		expense.Description,
		expense.Date,
		finance.TransactionTypeExpense,
		walletID,
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

	var walletID *finance.WalletID
	if income.WalletID != nil {
		wid := finance.NewWalletID(int(*income.WalletID))
		walletID = &wid
	}

	transaction := finance.NewTransaction(
		transactionID,
		userID,
		categoryID,
		currencyID,
		amount,
		income.IsApproved,
		income.Description,
		income.Date,
		finance.TransactionTypeIncome,
		walletID,
	)
	return transaction
}

// GetTotalCount gets the total count of transactions
func (r *GormTransactionRepository) GetTotalCount(ctx context.Context) (int, error) {
	var expenseCount, incomeCount int64

	// Count expenses
	err := r.db.WithContext(ctx).Model(&Expense{}).Count(&expenseCount).Error
	if err != nil {
		return 0, err
	}

	// Count incomes
	err = r.db.WithContext(ctx).Model(&Income{}).Count(&incomeCount).Error
	if err != nil {
		return 0, err
	}

	return int(expenseCount + incomeCount), nil
}

// GetTotalExpenses gets the total amount of expenses
func (r *GormTransactionRepository) GetTotalExpenses(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&Expense{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetTotalIncome gets the total amount of income
func (r *GormTransactionRepository) GetTotalIncome(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&Income{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

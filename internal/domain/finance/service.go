package finance

import (
	"context"
	"errors"
	"time"
)

// TransactionService handles transaction-related domain operations
type TransactionService struct {
	transactionRepo TransactionRepository
	categoryRepo    CategoryRepository
	currencyRepo    CurrencyRepository
	walletRepo      WalletRepository
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo TransactionRepository,
	categoryRepo CategoryRepository,
	currencyRepo CurrencyRepository,
	walletRepo WalletRepository,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		currencyRepo:    currencyRepo,
		walletRepo:      walletRepo,
	}
}

// CreateTransaction creates a new transaction
func (s *TransactionService) CreateTransaction(
	ctx context.Context,
	userID UserID,
	walletID WalletID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	date time.Time,
	transactionType TransactionType,
) (*Transaction, error) {
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return nil, errors.New("wallet not found")
	}
	if wallet.UserID().Value() != userID.Value() {
		return nil, errors.New("access denied to wallet")
	}
	if err := wallet.CanAcceptTransactions(); err != nil {
		return nil, err
	}
	if currencyID.Value() == 0 {
		currencyID = wallet.CurrencyID()
	}
	if wallet.CurrencyID().Value() != currencyID.Value() {
		return nil, errors.New("transaction currency must match wallet currency")
	}

	// Validate category exists and user has access
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Check if user has access to category (default or user's own)
	if !category.IsDefault() && (category.UserID() == nil || category.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to category")
	}

	// Validate category type matches transaction type
	if category.Type() != CategoryType(transactionType) {
		return nil, errors.New("category type does not match transaction type")
	}

	// Validate currency exists and user has access
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return nil, errors.New("currency not found")
	}

	// Check if user has access to currency (default or user's own)
	if !currency.IsDefault() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to currency")
	}

	money, err := NewMoney(amount.Amount(), currencyID)
	if err != nil {
		return nil, err
	}

	// Create transaction
	transaction := NewTransaction(
		TransactionID{}, // Will be set by repository
		userID,
		walletID,
		categoryID,
		currencyID,
		money,
		description,
		date,
		transactionType,
	)

	// Save transaction
	if err := s.transactionRepo.Save(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetTransactionsByUser retrieves all transactions for a user
func (s *TransactionService) GetTransactionsByUser(ctx context.Context, userID UserID) ([]*Transaction, error) {
	return s.transactionRepo.FindByUserID(ctx, userID)
}

// GetTransactionsByUserAndDateRange retrieves transactions for a user within a date range
func (s *TransactionService) GetTransactionsByUserAndDateRange(
	ctx context.Context,
	userID UserID,
	startDate, endDate time.Time,
) ([]*Transaction, error) {
	return s.transactionRepo.FindByUserIDAndDateRange(ctx, userID, startDate, endDate)
}

// GetTransactionsByUserWithFilters retrieves transactions for a user with filters
func (s *TransactionService) GetTransactionsByUserWithFilters(
	ctx context.Context,
	userID UserID,
	filters TransactionFilters,
) ([]*Transaction, int64, error) {
	return s.transactionRepo.FindByUserIDWithFilters(ctx, userID, filters)
}

// UpdateTransaction updates a transaction
func (s *TransactionService) UpdateTransaction(
	ctx context.Context,
	transactionID TransactionID,
	userID UserID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	date time.Time,
	expectedType TransactionType,
) (*Transaction, error) {
	// Get transaction to verify ownership, querying the correct table first based on expected type
	transaction, err := s.transactionRepo.FindByIDAndType(ctx, transactionID, expectedType)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if transaction.UserID().Value() != userID.Value() {
		return nil, errors.New("access denied")
	}

	// Verify transaction type matches expected type (double-check for safety)
	if transaction.Type() != expectedType {
		return nil, errors.New("transaction type mismatch")
	}

	// Validate category exists and user has access
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Check if user has access to category (default or user's own)
	if !category.IsDefault() && (category.UserID() == nil || category.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to category")
	}

	// Validate currency exists and user has access
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return nil, errors.New("currency not found")
	}

	// Check if user has access to currency (default or user's own)
	if !currency.IsDefault() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to currency")
	}

	// Update transaction fields
	if err := transaction.UpdateAmount(amount); err != nil {
		return nil, err
	}
	transaction.UpdateDescription(description)
	transaction.UpdateDate(date)

	// Update the category and currency IDs (these need to be set directly)
	transaction.categoryID = categoryID
	transaction.currencyID = currencyID

	// Save updated transaction
	if err := s.transactionRepo.Save(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// DeleteTransaction deletes a transaction
func (s *TransactionService) DeleteTransaction(ctx context.Context, transactionID TransactionID, userID UserID) error {
	// Get transaction to verify ownership
	transaction, err := s.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return errors.New("transaction not found")
	}

	if transaction.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	return s.transactionRepo.Delete(ctx, transactionID)
}

// CategoryService handles category-related domain operations
type CategoryService struct {
	categoryRepo CategoryRepository
	budgetRepo   BudgetRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo CategoryRepository, budgetRepo BudgetRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		budgetRepo:   budgetRepo,
	}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(
	ctx context.Context,
	userID UserID,
	name string,
	color string,
	categoryType CategoryType,
) (*Category, error) {
	// Create category
	category, err := NewCategory(
		CategoryID{}, // Will be set by repository
		&userID,
		name,
		color,
		false, // User categories are not default
		categoryType,
	)
	if err != nil {
		return nil, err
	}

	// Save category
	if err := s.categoryRepo.Save(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategoriesByUser retrieves all categories accessible to a user
func (s *CategoryService) GetCategoriesByUser(ctx context.Context, userID UserID) ([]*Category, error) {
	// Get user's categories
	userCategories, err := s.categoryRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get default categories
	defaultCategories, err := s.categoryRepo.FindDefaultCategories(ctx)
	if err != nil {
		return nil, err
	}

	// Combine and return
	allCategories := append(defaultCategories, userCategories...)
	return allCategories, nil
}

// GetCategoriesByUserAndType retrieves categories by user and type
func (s *CategoryService) GetCategoriesByUserAndType(
	ctx context.Context,
	userID UserID,
	categoryType CategoryType,
) ([]*Category, error) {
	return s.categoryRepo.FindByUserIDAndType(ctx, userID, categoryType)
}

// GetCategoryByID retrieves a category by ID
func (s *CategoryService) GetCategoryByID(ctx context.Context, categoryID CategoryID) (*Category, error) {
	return s.categoryRepo.FindByID(ctx, categoryID)
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(
	ctx context.Context,
	categoryID CategoryID,
	userID UserID,
	name string,
	color string,
	categoryType CategoryType,
) error {
	// Get category
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return errors.New("category not found")
	}

	// Check if user can update this category
	if category.IsDefault() {
		return errors.New("cannot update default category")
	}

	if category.UserID() == nil || category.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	// Update category
	if err := category.UpdateName(name); err != nil {
		return err
	}

	category.UpdateColor(color)

	// Save updated category
	return s.categoryRepo.Save(ctx, category)
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID CategoryID, userID UserID) error {
	// Get category
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return errors.New("category not found")
	}

	// Check if user can delete this category
	if !category.CanBeDeleted() {
		return errors.New("cannot delete default category")
	}

	if category.UserID() == nil || category.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	existingBudgets, err := s.budgetRepo.FindByUserIDAndCategory(ctx, userID, categoryID)
	if err != nil {
		return err
	}
	if len(existingBudgets) > 0 {
		return errors.New("cannot delete category with existing budgets")
	}

	return s.categoryRepo.Delete(ctx, categoryID)
}

// BudgetService handles budget-related domain operations
type BudgetService struct {
	budgetRepo   BudgetRepository
	categoryRepo CategoryRepository
}

// NewBudgetService creates a new budget service
func NewBudgetService(budgetRepo BudgetRepository, categoryRepo CategoryRepository) *BudgetService {
	return &BudgetService{
		budgetRepo:   budgetRepo,
		categoryRepo: categoryRepo,
	}
}

// GetBudgetByID retrieves a budget by its ID
func (s *BudgetService) GetBudgetByID(ctx context.Context, budgetID BudgetID) (*Budget, error) {
	return s.budgetRepo.FindByID(ctx, budgetID)
}

// CreateBudget creates a new budget
func (s *BudgetService) CreateBudget(
	ctx context.Context,
	userID UserID,
	categoryID CategoryID,
	amount Money,
	limitType BudgetLimitType,
	percent *float64,
	period BudgetPeriod,
	startDate time.Time,
) (*Budget, error) {
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	if !category.IsDefault() && (category.UserID() == nil || category.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to category")
	}

	if category.Type() != CategoryTypeExpense {
		return nil, errors.New("budget category must be expense type")
	}

	budget, err := NewBudget(
		BudgetID{},
		userID,
		categoryID,
		amount,
		limitType,
		percent,
		period,
		startDate,
	)
	if err != nil {
		return nil, err
	}

	if err := s.ensureNoOverlap(ctx, userID, categoryID, budget.StartDate(), budget.EndDate(), BudgetID{}); err != nil {
		return nil, err
	}

	if err := s.budgetRepo.Save(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

// GetBudgetsByUser retrieves all budgets for a user
func (s *BudgetService) GetBudgetsByUser(ctx context.Context, userID UserID) ([]*Budget, error) {
	return s.budgetRepo.FindByUserID(ctx, userID)
}

// GetActiveBudgetsByUser retrieves active budgets for a user
func (s *BudgetService) GetActiveBudgetsByUser(ctx context.Context, userID UserID) ([]*Budget, error) {
	return s.budgetRepo.FindActiveByUserID(ctx, userID)
}

func (s *BudgetService) ensureNoOverlap(
	ctx context.Context,
	userID UserID,
	categoryID CategoryID,
	startDate, endDate time.Time,
	excludeID BudgetID,
) error {
	existing, err := s.budgetRepo.FindByUserIDAndCategory(ctx, userID, categoryID)
	if err != nil {
		return err
	}

	for _, candidate := range existing {
		if excludeID.Value() != 0 && candidate.ID().Value() == excludeID.Value() {
			continue
		}
		if candidate.OverlapsWith(startDate, endDate) {
			return errors.New("overlapping budget already exists for this category")
		}
	}

	return nil
}

// UpdateBudget updates a budget
func (s *BudgetService) UpdateBudget(
	ctx context.Context,
	budgetID BudgetID,
	userID UserID,
	categoryID CategoryID,
	amount Money,
	limitType BudgetLimitType,
	percent *float64,
	period BudgetPeriod,
	startDate time.Time,
	endDate time.Time,
) (*Budget, error) {
	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, errors.New("budget not found")
	}

	if budget.UserID().Value() != userID.Value() {
		return nil, errors.New("budget not found")
	}

	targetCategoryID := budget.CategoryID()
	if categoryID.Value() != 0 {
		category, err := s.categoryRepo.FindByID(ctx, categoryID)
		if err != nil {
			return nil, errors.New("category not found")
		}
		if !category.IsDefault() && (category.UserID() == nil || category.UserID().Value() != userID.Value()) {
			return nil, errors.New("access denied to category")
		}
		if category.Type() != CategoryTypeExpense {
			return nil, errors.New("budget category must be expense type")
		}
		budget.categoryID = categoryID
		targetCategoryID = categoryID
	}

	if err := budget.UpdateLimit(limitType, amount, percent); err != nil {
		return nil, err
	}

	if err := budget.UpdatePeriod(period); err != nil {
		return nil, err
	}

	if err := budget.UpdateStartDate(startDate); err != nil {
		return nil, err
	}

	if err := budget.UpdateEndDate(endDate); err != nil {
		return nil, err
	}

	if err := s.ensureNoOverlap(ctx, userID, targetCategoryID, budget.StartDate(), budget.EndDate(), budget.ID()); err != nil {
		return nil, err
	}

	if err := s.budgetRepo.Save(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

// DeleteBudget deletes a budget
func (s *BudgetService) DeleteBudget(ctx context.Context, budgetID BudgetID, userID UserID) error {
	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return errors.New("budget not found")
	}

	if budget.UserID().Value() != userID.Value() {
		return errors.New("budget not found")
	}

	if err := s.budgetRepo.DeleteByIDAndUserID(ctx, budgetID, userID); err != nil {
		return errors.New("budget not found")
	}

	return nil
}

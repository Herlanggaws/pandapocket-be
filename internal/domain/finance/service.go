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
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo TransactionRepository,
	categoryRepo CategoryRepository,
	currencyRepo CurrencyRepository,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		currencyRepo:    currencyRepo,
	}
}

// CreateTransaction creates a new transaction
func (s *TransactionService) CreateTransaction(
	ctx context.Context,
	userID UserID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	date time.Time,
	transactionType TransactionType,
) (*Transaction, error) {
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

	// Create transaction
	transaction := NewTransaction(
		TransactionID{}, // Will be set by repository
		userID,
		categoryID,
		currencyID,
		amount,
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
) (*Transaction, error) {
	// Get transaction to verify ownership
	transaction, err := s.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if transaction.UserID().Value() != userID.Value() {
		return nil, errors.New("access denied")
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
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
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

// CreateBudget creates a new budget
func (s *BudgetService) CreateBudget(
	ctx context.Context,
	userID UserID,
	categoryID CategoryID,
	amount Money,
	period BudgetPeriod,
	startDate time.Time,
) (*Budget, error) {
	// Validate category exists and user has access
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Check if user has access to category (default or user's own)
	if !category.IsDefault() && (category.UserID() == nil || category.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to category")
	}

	// Create budget
	budget, err := NewBudget(
		BudgetID{}, // Will be set by repository
		userID,
		categoryID,
		amount,
		period,
		startDate,
	)
	if err != nil {
		return nil, err
	}

	// Save budget
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

// UpdateBudget updates a budget
func (s *BudgetService) UpdateBudget(
	ctx context.Context,
	budgetID BudgetID,
	userID UserID,
	amount Money,
	period BudgetPeriod,
	startDate time.Time,
) error {
	// Get budget
	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return errors.New("budget not found")
	}

	// Check if user can update this budget
	if budget.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	// Update budget
	if err := budget.UpdateAmount(amount); err != nil {
		return err
	}

	if err := budget.UpdatePeriod(period); err != nil {
		return err
	}

	budget.UpdateStartDate(startDate)

	// Save updated budget
	return s.budgetRepo.Save(ctx, budget)
}

// DeleteBudget deletes a budget
func (s *BudgetService) DeleteBudget(ctx context.Context, budgetID BudgetID, userID UserID) error {
	// Get budget to verify ownership
	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return errors.New("budget not found")
	}

	if budget.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	return s.budgetRepo.Delete(ctx, budgetID)
}

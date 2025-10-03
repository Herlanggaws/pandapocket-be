package identity

import (
	"context"
	domainIdentity "panda-pocket/internal/domain/identity"
	"time"
)

// DashboardStatsResponse represents the dashboard statistics response
type DashboardStatsResponse struct {
	TotalUsers              int     `json:"total_users"`
	ActiveUsers             int     `json:"active_users"`
	TotalBudgets            int     `json:"total_budgets"`
	TotalTransactions       int     `json:"total_transactions"`
	TotalExpenses           float64 `json:"total_expenses"`
	TotalIncome             float64 `json:"total_income"`
	BudgetsCreatedThisWeek  int     `json:"budgets_created_this_week"`
	BudgetsCreatedThisMonth int     `json:"budgets_created_this_month"`
}

// GetDashboardStatsUseCase handles getting dashboard statistics
type GetDashboardStatsUseCase struct {
	userRepo        domainIdentity.UserRepository
	budgetRepo      BudgetRepository
	transactionRepo TransactionRepository
}

// BudgetRepository defines the contract for budget persistence
type BudgetRepository interface {
	GetTotalCount(ctx context.Context) (int, error)
	GetCountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
}

// TransactionRepository defines the contract for transaction persistence
type TransactionRepository interface {
	GetTotalCount(ctx context.Context) (int, error)
	GetTotalExpenses(ctx context.Context) (float64, error)
	GetTotalIncome(ctx context.Context) (float64, error)
}

// NewGetDashboardStatsUseCase creates a new get dashboard stats use case
func NewGetDashboardStatsUseCase(
	userRepo domainIdentity.UserRepository,
	budgetRepo BudgetRepository,
	transactionRepo TransactionRepository,
) *GetDashboardStatsUseCase {
	return &GetDashboardStatsUseCase{
		userRepo:        userRepo,
		budgetRepo:      budgetRepo,
		transactionRepo: transactionRepo,
	}
}

// Execute executes the get dashboard stats use case
func (uc *GetDashboardStatsUseCase) Execute(ctx context.Context) (*DashboardStatsResponse, error) {
	// Get total users (only users with "user" role)
	users, err := uc.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Filter users by role - only count users with "user" role
	var userRoleUsers []*domainIdentity.User
	for _, user := range users {
		if user.Role().Value() == "user" {
			userRoleUsers = append(userRoleUsers, user)
		}
	}
	totalUsers := len(userRoleUsers)

	// Get active users (last 30 days)
	// Note: This is a simplified implementation. In a real implementation,
	// you'd want to add last_login_at to the domain user and repository
	// For now, we'll use a placeholder logic - count all users as active
	activeUsers := len(userRoleUsers)

	// Get total budgets
	totalBudgets, err := uc.budgetRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, err
	}

	// Get total transactions
	totalTransactions, err := uc.transactionRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, err
	}

	// Get total expenses
	totalExpenses, err := uc.transactionRepo.GetTotalExpenses(ctx)
	if err != nil {
		return nil, err
	}

	// Get total income
	totalIncome, err := uc.transactionRepo.GetTotalIncome(ctx)
	if err != nil {
		return nil, err
	}

	// Get budgets created this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	budgetsThisWeek, err := uc.budgetRepo.GetCountByDateRange(ctx, weekAgo, time.Now())
	if err != nil {
		return nil, err
	}

	// Get budgets created this month
	monthAgo := time.Now().AddDate(0, 0, -30)
	budgetsThisMonth, err := uc.budgetRepo.GetCountByDateRange(ctx, monthAgo, time.Now())
	if err != nil {
		return nil, err
	}

	return &DashboardStatsResponse{
		TotalUsers:              totalUsers,
		ActiveUsers:             activeUsers,
		TotalBudgets:            totalBudgets,
		TotalTransactions:       totalTransactions,
		TotalExpenses:           totalExpenses,
		TotalIncome:             totalIncome,
		BudgetsCreatedThisWeek:  budgetsThisWeek,
		BudgetsCreatedThisMonth: budgetsThisMonth,
	}, nil
}

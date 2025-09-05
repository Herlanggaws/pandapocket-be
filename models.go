package main

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Currency struct {
	ID        int       `json:"id"`
	UserID    *int      `json:"user_id,omitempty"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Symbol    string    `json:"symbol"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID           int       `json:"id"`
	UserID       *int      `json:"user_id,omitempty"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
	IsDefault    bool      `json:"is_default"`
	CategoryType string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
}

type Expense struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CategoryID  int       `json:"category_id"`
	Category    Category  `json:"category"`
	CurrencyID  int       `json:"currency_id"`
	Currency    Currency  `json:"currency"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}

type Income struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CategoryID  int       `json:"category_id"`
	Category    Category  `json:"category"`
	CurrencyID  int       `json:"currency_id"`
	Currency    Currency  `json:"currency"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}

type Budget struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	CategoryID int       `json:"category_id"`
	Category   Category  `json:"category"`
	Amount     float64   `json:"amount"`
	Period     string    `json:"period"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	CreatedAt  time.Time `json:"created_at"`
	// Computed fields
	SpentAmount float64 `json:"spent_amount,omitempty"`
	Remaining   float64 `json:"remaining,omitempty"`
	Progress    float64 `json:"progress,omitempty"`
}

type RecurringTransaction struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	CategoryID   int       `json:"category_id"`
	Category     Category  `json:"category"`
	CurrencyID   int       `json:"currency_id"`
	Currency     Currency  `json:"currency"`
	Amount       float64   `json:"amount"`
	Description  string    `json:"description"`
	Frequency    string    `json:"frequency"`
	NextDueDate  time.Time `json:"next_due_date"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserPreferences struct {
	ID                   int       `json:"id"`
	UserID               int       `json:"user_id"`
	PrimaryCurrencyID    int       `json:"primary_currency_id"`
	PrimaryCurrency      Currency  `json:"primary_currency"`
	EmailNotifications   bool      `json:"email_notifications"`
	BudgetAlerts         bool      `json:"budget_alerts"`
	RecurringReminders   bool      `json:"recurring_reminders"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/Response types
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
	Type  string `json:"type" binding:"required,oneof=expense income"`
}

type CreateExpenseRequest struct {
	CategoryID  int     `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
}

type CreateIncomeRequest struct {
	CategoryID  int     `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
}

type CreateCurrencyRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Symbol string `json:"symbol" binding:"required"`
}

type UpdateCurrencyRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Symbol string `json:"symbol" binding:"required"`
}

type CreateBudgetRequest struct {
	CategoryID int     `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Period     string  `json:"period" binding:"required,oneof=weekly monthly yearly"`
	StartDate  string  `json:"start_date" binding:"required"`
}

type CreateRecurringTransactionRequest struct {
	CategoryID  int     `json:"category_id" binding:"required"`
	CurrencyID  int     `json:"currency_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Frequency   string  `json:"frequency" binding:"required,oneof=daily weekly monthly yearly"`
	NextDueDate string  `json:"next_due_date" binding:"required"`
}

type UpdateUserPreferencesRequest struct {
	PrimaryCurrencyID  int  `json:"primary_currency_id" binding:"required"`
	EmailNotifications bool `json:"email_notifications"`
	BudgetAlerts       bool `json:"budget_alerts"`
	RecurringReminders bool `json:"recurring_reminders"`
}

type DashboardResponse struct {
	RecentExpenses []Expense  `json:"recent_expenses"`
	RecentIncomes  []Income   `json:"recent_incomes"`
	MonthlyTotal   float64    `json:"monthly_total"`
	MonthlyIncome  float64    `json:"monthly_income"`
	Balance        float64    `json:"balance"`
	Categories     []Category `json:"categories"`
}

// Analytics types
type SpendingByCategory struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategoryColor string `json:"category_color"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

type SpendingByPeriod struct {
	Period string  `json:"period"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

type AnalyticsResponse struct {
	SpendingByCategory []SpendingByCategory `json:"spending_by_category"`
	SpendingByPeriod   []SpendingByPeriod   `json:"spending_by_period"`
	TotalSpent         float64              `json:"total_spent"`
	TotalIncome        float64              `json:"total_income"`
	NetAmount          float64              `json:"net_amount"`
}

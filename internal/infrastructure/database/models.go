package database

import (
	"time"
)

// User represents a user in the database
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Currencies            []Currency             `gorm:"foreignKey:UserID" json:"currencies,omitempty"`
	Categories            []Category             `gorm:"foreignKey:UserID" json:"categories,omitempty"`
	Expenses              []Expense              `gorm:"foreignKey:UserID" json:"expenses,omitempty"`
	Incomes               []Income               `gorm:"foreignKey:UserID" json:"incomes,omitempty"`
	Budgets               []Budget               `gorm:"foreignKey:UserID" json:"budgets,omitempty"`
	RecurringTransactions []RecurringTransaction `gorm:"foreignKey:UserID" json:"recurring_transactions,omitempty"`
	UserPreferences       *UserPreferences       `gorm:"foreignKey:UserID" json:"user_preferences,omitempty"`
	Notifications         []Notification         `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
}

// Currency represents a currency in the database
type Currency struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `gorm:"index" json:"user_id,omitempty"`
	Code      string    `gorm:"not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	Symbol    string    `gorm:"not null" json:"symbol"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User                  *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Expenses              []Expense              `gorm:"foreignKey:CurrencyID" json:"expenses,omitempty"`
	Incomes               []Income               `gorm:"foreignKey:CurrencyID" json:"incomes,omitempty"`
	RecurringTransactions []RecurringTransaction `gorm:"foreignKey:CurrencyID" json:"recurring_transactions,omitempty"`
	UserPreferences       []UserPreferences      `gorm:"foreignKey:PrimaryCurrencyID" json:"user_preferences,omitempty"`
}

// Category represents a category in the database
type Category struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `gorm:"index" json:"user_id,omitempty"`
	Name         string    `gorm:"not null" json:"name"`
	Color        string    `gorm:"default:'#3B82F6'" json:"color"`
	IsDefault    bool      `gorm:"default:false" json:"is_default"`
	CategoryType string    `gorm:"default:'expense'" json:"category_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User                  *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Expenses              []Expense              `gorm:"foreignKey:CategoryID" json:"expenses,omitempty"`
	Incomes               []Income               `gorm:"foreignKey:CategoryID" json:"incomes,omitempty"`
	Budgets               []Budget               `gorm:"foreignKey:CategoryID" json:"budgets,omitempty"`
	RecurringTransactions []RecurringTransaction `gorm:"foreignKey:CategoryID" json:"recurring_transactions,omitempty"`
}

// Expense represents an expense transaction in the database
type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID  uint      `gorm:"not null;index" json:"currency_id"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Income represents an income transaction in the database
type Income struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID  uint      `gorm:"not null;index" json:"currency_id"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Budget represents a budget in the database
type Budget struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	CategoryID uint      `gorm:"not null;index" json:"category_id"`
	Amount     float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Period     string    `gorm:"not null;check:period IN ('weekly', 'monthly', 'yearly')" json:"period"`
	StartDate  time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:date;not null" json:"end_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// RecurringTransaction represents a recurring transaction in the database
type RecurringTransaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID  uint      `gorm:"not null;index" json:"currency_id"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Frequency   string    `gorm:"not null;check:frequency IN ('daily', 'weekly', 'monthly', 'yearly')" json:"frequency"`
	NextDueDate time.Time `gorm:"type:date;not null" json:"next_due_date"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// UserPreferences represents user preferences in the database
type UserPreferences struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserID             uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	PrimaryCurrencyID  uint      `gorm:"not null" json:"primary_currency_id"`
	EmailNotifications bool      `gorm:"default:true" json:"email_notifications"`
	BudgetAlerts       bool      `gorm:"default:true" json:"budget_alerts"`
	RecurringReminders bool      `gorm:"default:true" json:"recurring_reminders"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relationships
	User            *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PrimaryCurrency *Currency `gorm:"foreignKey:PrimaryCurrencyID" json:"primary_currency,omitempty"`
}

// Notification represents a notification in the database
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Title     string    `gorm:"not null" json:"title"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Type      string    `gorm:"not null" json:"type"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName methods for custom table names (optional)
func (User) TableName() string {
	return "users"
}

func (Currency) TableName() string {
	return "currencies"
}

func (Category) TableName() string {
	return "categories"
}

func (Expense) TableName() string {
	return "expenses"
}

func (Income) TableName() string {
	return "incomes"
}

func (Budget) TableName() string {
	return "budgets"
}

func (RecurringTransaction) TableName() string {
	return "recurring_transactions"
}

func (UserPreferences) TableName() string {
	return "user_preferences"
}

func (Notification) TableName() string {
	return "notifications"
}

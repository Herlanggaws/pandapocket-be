package database

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         string     `gorm:"default:'user';check:role IN ('user', 'admin', 'super_admin')" json:"role"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	Currencies            []Currency             `gorm:"foreignKey:UserID" json:"currencies,omitempty"`
	Categories            []Category             `gorm:"foreignKey:UserID" json:"categories,omitempty"`
	Wallets               []Wallet               `gorm:"foreignKey:UserID" json:"wallets,omitempty"`
	Expenses              []Expense              `gorm:"foreignKey:UserID" json:"expenses,omitempty"`
	Incomes               []Income               `gorm:"foreignKey:UserID" json:"incomes,omitempty"`
	Budgets               []Budget               `gorm:"foreignKey:UserID" json:"budgets,omitempty"`
	RecurringTransactions []RecurringTransaction `gorm:"foreignKey:UserID" json:"recurring_transactions,omitempty"`
	Transfers             []Transfer             `gorm:"foreignKey:UserID" json:"transfers,omitempty"`
	UserPreferences       *UserPreferences       `gorm:"foreignKey:UserID" json:"user_preferences,omitempty"`
	Subscription          *Subscription          `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
	Notifications         []Notification         `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
}

// Wallet represents a money account / dompet for a user
type Wallet struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	Name           string    `gorm:"not null" json:"name"`
	Type           string    `gorm:"not null;default:'cash';check:type IN ('cash', 'bank', 'e_wallet')" json:"type"`
	CurrencyID     uint      `gorm:"not null;index" json:"currency_id"`
	OpeningBalance float64   `gorm:"type:decimal(12,2);not null;default:0" json:"opening_balance"`
	IsDefault      bool      `gorm:"default:false;index" json:"is_default"`
	IsArchived     bool      `gorm:"default:false;index" json:"is_archived"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Transfer represents a same-currency move between two wallets
type Transfer struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	FromWalletID uint      `gorm:"not null;index" json:"from_wallet_id"`
	ToWalletID   uint      `gorm:"not null;index" json:"to_wallet_id"`
	Amount       float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	Description  string    `gorm:"type:text" json:"description"`
	Date         time.Time `gorm:"type:date;not null;index" json:"date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	User       *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FromWallet *Wallet `gorm:"foreignKey:FromWalletID" json:"from_wallet,omitempty"`
	ToWallet   *Wallet `gorm:"foreignKey:ToWalletID" json:"to_wallet,omitempty"`
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
	UserID      uint      `gorm:"not null;index;index:idx_expense_user_date_created,priority:1" json:"user_id"`
	WalletID    *uint     `gorm:"index" json:"wallet_id,omitempty"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID  uint      `gorm:"not null;index" json:"currency_id"`
	Amount      float64   `gorm:"type:decimal(14,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Date        time.Time `gorm:"type:date;not null;index:idx_expense_user_date_created,priority:2" json:"date"`
	CreatedAt   time.Time `gorm:"index:idx_expense_user_date_created,priority:3" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wallet   *Wallet   `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Income represents an income transaction in the database
type Income struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index;index:idx_income_user_date_created,priority:1" json:"user_id"`
	WalletID    *uint     `gorm:"index" json:"wallet_id,omitempty"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID  uint      `gorm:"not null;index" json:"currency_id"`
	Amount      float64   `gorm:"type:decimal(14,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Date        time.Time `gorm:"type:date;not null;index:idx_income_user_date_created,priority:2" json:"date"`
	CreatedAt   time.Time `gorm:"index:idx_income_user_date_created,priority:3" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wallet   *Wallet   `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Budget represents a budget in the database
type Budget struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	CategoryID uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID uint      `gorm:"not null;index;default:1" json:"currency_id"`
	Amount     float64   `gorm:"type:decimal(14,2);not null;default:0" json:"amount"`
	LimitType  string    `gorm:"not null;default:'fixed';check:limit_type IN ('fixed', 'percent')" json:"limit_type"`
	Percent    *float64  `gorm:"type:decimal(5,2)" json:"percent,omitempty"`
	Period     string    `gorm:"not null;check:period IN ('weekly', 'monthly', 'yearly')" json:"period"`
	StartDate  time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:date;not null" json:"end_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// FinancialGoal represents a savings goal with a deadline
type FinancialGoal struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Name          string    `gorm:"not null" json:"name"`
	TargetAmount  float64   `gorm:"type:decimal(14,2);not null" json:"target_amount"`
	CurrencyID    uint      `gorm:"not null;index" json:"currency_id"`
	CurrentAmount float64   `gorm:"type:decimal(14,2);not null;default:0" json:"current_amount"`
	TargetDate    time.Time `gorm:"type:date;not null" json:"target_date"`
	Status        string    `gorm:"not null;default:'active';check:status IN ('active', 'completed', 'archived')" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Asset represents a non-wallet asset position
type Asset struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	Name         string     `gorm:"not null" json:"name"`
	Type         string     `gorm:"not null;default:'other';check:type IN ('property', 'vehicle', 'investment', 'other')" json:"type"`
	CurrencyID   uint       `gorm:"not null;index" json:"currency_id"`
	CurrentValue float64    `gorm:"type:decimal(14,2);not null;default:0" json:"current_value"`
	Notes        string     `gorm:"type:text" json:"notes"`
	IsArchived   bool       `gorm:"default:false;index" json:"is_archived"`
	AsOfDate     *time.Time `gorm:"type:date" json:"as_of_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Liability represents an amount owed
type Liability struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	UserID            uint       `gorm:"not null;index" json:"user_id"`
	Name              string     `gorm:"not null" json:"name"`
	Type              string     `gorm:"not null;default:'other';check:type IN ('loan', 'credit_card', 'mortgage', 'other')" json:"type"`
	CurrencyID        uint       `gorm:"not null;index" json:"currency_id"`
	CurrentBalance    float64    `gorm:"type:decimal(14,2);not null;default:0" json:"current_balance"`
	OriginalPrincipal *float64   `gorm:"type:decimal(14,2)" json:"original_principal,omitempty"`
	InterestRateAPR   *float64   `gorm:"type:decimal(8,4)" json:"interest_rate_apr,omitempty"`
	MinimumPayment    float64    `gorm:"type:decimal(14,2);not null;default:0" json:"minimum_payment"`
	NextDueDate       *time.Time `gorm:"type:date" json:"next_due_date,omitempty"`
	Notes             string     `gorm:"type:text" json:"notes"`
	IsArchived        bool       `gorm:"default:false;index" json:"is_archived"`
	AsOfDate          *time.Time `gorm:"type:date" json:"as_of_date,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// LiabilityPayment records a payment against a liability
type LiabilityPayment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	LiabilityID uint      `gorm:"not null;index" json:"liability_id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Amount      float64   `gorm:"type:decimal(14,2);not null" json:"amount"`
	PaidAt      time.Time `gorm:"type:date;not null" json:"paid_at"`
	ExpenseID   *uint     `gorm:"index" json:"expense_id,omitempty"`
	Note        string    `gorm:"type:text" json:"note"`
	CreatedAt   time.Time `json:"created_at"`

	Liability *Liability `gorm:"foreignKey:LiabilityID" json:"liability,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (LiabilityPayment) TableName() string {
	return "liability_payments"
}

// HealthScoreSnapshot stores monthly health score history
type HealthScoreSnapshot struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null;uniqueIndex:idx_health_user_month" json:"user_id"`
	YearMonth       string    `gorm:"size:7;not null;uniqueIndex:idx_health_user_month" json:"year_month"`
	Score           int       `gorm:"not null" json:"score"`
	BudgetAdherence float64   `gorm:"type:decimal(6,2);not null" json:"budget_adherence"`
	Cashflow        float64   `gorm:"type:decimal(6,2);not null" json:"cashflow"`
	Coverage        float64   `gorm:"type:decimal(6,2);not null" json:"coverage"`
	ComputedAt      time.Time `gorm:"not null" json:"computed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// RecurringTransaction represents a recurring transaction in the database
type RecurringTransaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	WalletID    *uint     `gorm:"index" json:"wallet_id,omitempty"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	CurrencyID  uint      `gorm:"not null;index" json:"currency_id"`
	Amount      float64   `gorm:"type:decimal(14,2);not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Frequency   string    `gorm:"not null;check:frequency IN ('daily', 'weekly', 'monthly', 'yearly')" json:"frequency"`
	Type        string    `gorm:"not null;default:'expense';check:type IN ('expense', 'income')" json:"type"`
	Weekday     *int      `gorm:"column:weekday" json:"weekday,omitempty"`
	DayOfMonth  *int      `gorm:"column:day_of_month" json:"day_of_month,omitempty"`
	MonthOfYear *int      `gorm:"column:month_of_year" json:"month_of_year,omitempty"`
	NextDueDate time.Time `gorm:"type:date;not null" json:"next_due_date"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wallet   *Wallet   `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency *Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// PendingTransaction represents a due recurring occurrence awaiting confirmation
type PendingTransaction struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	UserID                 uint       `gorm:"not null;index" json:"user_id"`
	WalletID               *uint      `gorm:"index" json:"wallet_id,omitempty"`
	RecurringTransactionID uint       `gorm:"not null;uniqueIndex:idx_pending_recurring_due" json:"recurring_transaction_id"`
	DueDate                time.Time  `gorm:"type:date;not null;uniqueIndex:idx_pending_recurring_due" json:"due_date"`
	Amount                 float64    `gorm:"type:decimal(14,2);not null" json:"amount"`
	Description            string     `gorm:"type:text" json:"description"`
	Type                   string     `gorm:"not null;check:type IN ('expense', 'income')" json:"type"`
	CategoryID             uint       `gorm:"not null;index" json:"category_id"`
	CurrencyID             uint       `gorm:"not null;index" json:"currency_id"`
	Status                 string     `gorm:"not null;default:'pending';check:status IN ('pending', 'confirmed', 'rejected');index" json:"status"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	ResolvedAt             *time.Time `json:"resolved_at,omitempty"`

	User                 *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wallet               *Wallet               `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	RecurringTransaction *RecurringTransaction `gorm:"foreignKey:RecurringTransactionID" json:"recurring_transaction,omitempty"`
	Category             *Category             `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Currency             *Currency             `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
}

// Subscription is the current billing entitlement row for a user (one per user).
type Subscription struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	UserID             uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	Plan               string     `gorm:"type:varchar(20);not null;default:'free';check:plan IN ('free','pro')" json:"plan"`
	BillingInterval    *string    `gorm:"type:varchar(20);check:billing_interval IS NULL OR billing_interval IN ('monthly','yearly')" json:"billing_interval,omitempty"`
	Status             string     `gorm:"type:varchar(20);not null;default:'expired';check:status IN ('trialing','active','past_due','canceled','expired')" json:"status"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	GraceEndsAt        *time.Time `json:"grace_ends_at,omitempty"`
	DoitSubscriptionID *string    `gorm:"type:varchar(128)" json:"doit_subscription_id,omitempty"`
	DoitCustomerRef    string     `gorm:"type:varchar(64);not null" json:"doit_customer_ref"`
	CancelAtPeriodEnd  bool       `gorm:"not null;default:false" json:"cancel_at_period_end"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// BillingWebhookEvent stores received doit webhook events for dedup (B5).
type BillingWebhookEvent struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EventID    string    `gorm:"uniqueIndex;type:varchar(128);not null" json:"event_id"`
	Payload    string    `gorm:"type:text;not null" json:"payload"`
	ReceivedAt time.Time `gorm:"not null" json:"received_at"`
}

// UserPreferences represents user preferences in the database
type UserPreferences struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserID             uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	PrimaryCurrencyID  uint      `gorm:"not null" json:"primary_currency_id"`
	EmailNotifications bool      `gorm:"default:true" json:"email_notifications"`
	BudgetAlerts       bool      `gorm:"default:true" json:"budget_alerts"`
	RecurringReminders bool      `gorm:"default:true" json:"recurring_reminders"`
	Language           string    `gorm:"size:8;default:'id'" json:"language"`
	Onboarding         JSONRaw   `gorm:"type:text;default:'{}'" json:"onboarding"`
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

// UserFeedback represents user-submitted product feedback
type UserFeedback struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Category  string    `gorm:"not null;size:32" json:"category"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// SupportTicket represents a Pro support ticket
type SupportTicket struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Subject   string    `gorm:"not null;size:200" json:"subject"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	Category  string    `gorm:"not null;size:32;index" json:"category"`
	Priority  string    `gorm:"not null;size:16;default:'medium'" json:"priority"`
	Status    string    `gorm:"not null;size:32;index;default:'open'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Token represents a JWT token in the database for revocation
type Token struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	AccessToken  string    `gorm:"not null" json:"-"` // Storing for potential reference, though we assume stateless access tokens usually
	RefreshToken string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt    time.Time `gorm:"not null;index" json:"expires_at"`
	Revoked      bool      `gorm:"default:false;index" json:"revoked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

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

func (Wallet) TableName() string {
	return "wallets"
}

func (Transfer) TableName() string {
	return "transfers"
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

func (PendingTransaction) TableName() string {
	return "pending_transactions"
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func (BillingWebhookEvent) TableName() string {
	return "billing_webhook_events"
}

func (UserPreferences) TableName() string {
	return "user_preferences"
}

func (Notification) TableName() string {
	return "notifications"
}

func (UserFeedback) TableName() string {
	return "user_feedbacks"
}

func (SupportTicket) TableName() string {
	return "support_tickets"
}

func (Token) TableName() string {
	return "tokens"
}

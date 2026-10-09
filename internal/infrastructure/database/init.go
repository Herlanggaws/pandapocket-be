package database

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	domainIdentity "panda-pocket/internal/domain/identity"
)

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		// .env file doesn't exist, use system environment variables
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// InitDB initializes the PostgreSQL database connection using GORM
func InitDB() (*gorm.DB, error) {
	// Load environment variables from .env file
	loadEnvFile()

	db, err := initGormDB()
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = autoMigrate(db)
	if err != nil {
		return nil, err
	}

	// Create default data
	err = createDefaultData(db)
	if err != nil {
		return nil, err
	}

	err = backfillDefaultWallets(db)
	if err != nil {
		return nil, err
	}

	err = backfillFreeSubscriptions(db)
	if err != nil {
		return nil, err
	}

	err = backfillBudgetLimitTypes(db)
	if err != nil {
		return nil, err
	}

	log.Printf("Database initialized successfully with GORM and PostgreSQL")
	return db, nil
}

func initGormDB() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "herlangga.wicaksono")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "panda_pocket")

	var dsn string
	if password != "" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	} else {
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
			host, port, user, dbname)
	}

	logMode := logger.Info
	if os.Getenv("GIN_MODE") == "release" {
		logMode = logger.Warn
	}
	config := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logMode),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// autoMigrate runs GORM auto-migration for all models
func autoMigrate(db *gorm.DB) error {
	// Drop password_reset_tokens table if exists to allow schema change (dev only)
	// In production, this would be a proper migration script
	if db.Migrator().HasTable("password_reset_tokens") {
		// Check if ID column is integer
		var columnType string
		db.Raw("SELECT data_type FROM information_schema.columns WHERE table_name = 'password_reset_tokens' AND column_name = 'id'").Scan(&columnType)

		if columnType == "bigint" || columnType == "integer" {
			log.Println("Dropping password_reset_tokens table to migrate to UUID primary key")
			if err := db.Migrator().DropTable("password_reset_tokens"); err != nil {
				return err
			}
		}
	}

	if db.Migrator().HasTable("ai_advisor_threads") {
		// Multi-thread: drop legacy one-thread-per-user unique index/constraint if present.
		_ = db.Exec(`ALTER TABLE ai_advisor_threads DROP CONSTRAINT IF EXISTS uni_ai_advisor_threads_user_id`).Error
		_ = db.Exec(`DROP INDEX IF EXISTS idx_ai_advisor_threads_user_id`).Error
		_ = db.Exec(`DROP INDEX IF EXISTS uni_ai_advisor_threads_user_id`).Error
	}

	if err := db.AutoMigrate(
		&User{},
		&Currency{},
		&Category{},
		&Wallet{},
		&Expense{},
		&Income{},
		&Budget{},
		&FinancialGoal{},
		&Asset{},
		&Liability{},
		&LiabilityPayment{},
		&Receivable{},
		&ReceivableCollection{},
		&GoalContribution{},
		&HealthScoreSnapshot{},
		&LoggingStreakFreeze{},
		&GoalMilestone{},
		&RecurringTransaction{},
		&PendingTransaction{},
		&Transfer{},
		&UserPreferences{},
		&Subscription{},
		&BillingWebhookEvent{},
		&PendingPayment{},
		&AICreditBalance{},
		&AICreditLedgerEntry{},
		&AIAdvisorThread{},
		&AIAdvisorMessage{},
		&Notification{},
		&UserFeedback{},
		&SupportTicket{},
		&FxRate{},
		&PasswordResetToken{},
		&AccountResetChallenge{},
		&Token{},
		&McpToken{},
	); err != nil {
		return err
	}
	if err := hashStoredSessionTokens(db); err != nil {
		return err
	}
	if err := backfillPublicIDs(db); err != nil {
		return err
	}
	return widenBillingIntervalCheck(db)
}

// backfillPublicIDs fills users that predate the external id. Existing values stay.
func backfillPublicIDs(db *gorm.DB) error {
	if !db.Migrator().HasTable(&User{}) {
		return nil
	}
	if err := fillBlankPublicIDs(db); err != nil {
		return err
	}
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_public_id ON users (public_id)`).Error; err != nil {
		return err
	}
	return db.Exec(`ALTER TABLE users ALTER COLUMN public_id SET NOT NULL`).Error
}

func fillBlankPublicIDs(db *gorm.DB) error {
	var users []User
	if err := db.Where("public_id IS NULL OR public_id = ''").Find(&users).Error; err != nil {
		return err
	}
	for _, user := range users {
		err := db.Model(&User{}).
			Where("id = ? AND (public_id IS NULL OR public_id = '')", user.ID).
			Update("public_id", uuid.NewString()).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// widenBillingIntervalCheck adds semiannual to the existing check.
// AutoMigrate does not rewrite a CHECK that is already on the table, and this statement does not update subscription rows.
func widenBillingIntervalCheck(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" || !db.Migrator().HasTable("subscriptions") {
		return nil
	}
	if err := db.Exec(`ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS chk_subscriptions_billing_interval`).Error; err != nil {
		return err
	}
	return db.Exec(`ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_billing_interval CHECK (billing_interval IS NULL OR billing_interval IN ('monthly','semiannual','yearly'))`).Error
}

// hashStoredSessionTokens replaces raw session JWTs with SHA-256 so a database dump is not a usable session.
// Rows that are already hashes are left unchanged.
func hashStoredSessionTokens(db *gorm.DB) error {
	if !db.Migrator().HasTable("tokens") {
		return nil
	}
	var rows []Token
	if err := db.Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if domainIdentity.IsSessionTokenHash(row.AccessToken) && domainIdentity.IsSessionTokenHash(row.RefreshToken) {
			continue
		}
		accessToken := row.AccessToken
		refreshToken := row.RefreshToken
		if !domainIdentity.IsSessionTokenHash(accessToken) {
			accessToken = domainIdentity.SessionTokenHash(accessToken)
		}
		if !domainIdentity.IsSessionTokenHash(refreshToken) {
			refreshToken = domainIdentity.SessionTokenHash(refreshToken)
		}
		err := db.Model(&Token{}).Where("id = ?", row.ID).Updates(map[string]any{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// createDefaultData creates default categories and currencies using GORM
func createDefaultData(db *gorm.DB) error {
	// Create default categories
	err := createDefaultCategoriesGorm(db)
	if err != nil {
		return err
	}

	if err := ensureDebtCategory(db); err != nil {
		return err
	}
	if err := ensureReceivableCategories(db); err != nil {
		return err
	}

	// Create default currencies
	err = createDefaultCurrenciesGorm(db)
	if err != nil {
		return err
	}

	return nil
}

func ensureDebtCategory(db *gorm.DB) error {
	var existing Category
	err := db.Where("name = ? AND is_default = ? AND category_type = ?", "Debt", true, "expense").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&Category{Name: "Debt", Color: "#DC2626", IsDefault: true, CategoryType: "expense"}).Error
}

func ensureReceivableCategories(db *gorm.DB) error {
	if err := ensureNamedDefaultCategory(db, "Receivable", "expense", "#0D9488"); err != nil {
		return err
	}
	return ensureNamedDefaultCategory(db, "Receivable", "income", "#0D9488")
}

func ensureNamedDefaultCategory(db *gorm.DB, name, categoryType, color string) error {
	var existing Category
	err := db.Where("name = ? AND is_default = ? AND category_type = ?", name, true, categoryType).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&Category{Name: name, Color: color, IsDefault: true, CategoryType: categoryType}).Error
}

// createDefaultCategoriesGorm creates default categories using GORM
func createDefaultCategoriesGorm(db *gorm.DB) error {
	// Check if default categories already exist
	var count int64
	err := db.Model(&Category{}).Where("is_default = ?", true).Count(&count).Error
	if err != nil {
		return err
	}

	// If default categories already exist, don't create them again
	if count > 0 {
		log.Println("Default categories already exist, skipping creation")
		return nil
	}

	// Default expense categories
	defaultExpenseCategories := []Category{
		{Name: "Food", Color: "#EF4444", IsDefault: true, CategoryType: "expense"},
		{Name: "Transport", Color: "#3B82F6", IsDefault: true, CategoryType: "expense"},
		{Name: "Entertainment", Color: "#8B5CF6", IsDefault: true, CategoryType: "expense"},
		{Name: "Shopping", Color: "#F59E0B", IsDefault: true, CategoryType: "expense"},
		{Name: "Bills", Color: "#10B981", IsDefault: true, CategoryType: "expense"},
		{Name: "Healthcare", Color: "#EC4899", IsDefault: true, CategoryType: "expense"},
		{Name: "Education", Color: "#06B6D4", IsDefault: true, CategoryType: "expense"},
		{Name: "Debt", Color: "#DC2626", IsDefault: true, CategoryType: "expense"},
		{Name: "Other", Color: "#6B7280", IsDefault: true, CategoryType: "expense"},
	}

	// Default income categories
	defaultIncomeCategories := []Category{
		{Name: "Salary", Color: "#10B981", IsDefault: true, CategoryType: "income"},
		{Name: "Bonus", Color: "#F59E0B", IsDefault: true, CategoryType: "income"},
		{Name: "Freelance", Color: "#8B5CF6", IsDefault: true, CategoryType: "income"},
		{Name: "Other", Color: "#6B7280", IsDefault: true, CategoryType: "income"},
	}

	// Create expense categories
	if err := db.Create(&defaultExpenseCategories).Error; err != nil {
		return err
	}

	// Create income categories
	if err := db.Create(&defaultIncomeCategories).Error; err != nil {
		return err
	}

	log.Printf("Created %d default expense categories and %d default income categories",
		len(defaultExpenseCategories), len(defaultIncomeCategories))
	return nil
}

// createDefaultCurrenciesGorm ensures system default currencies exist (insert missing by code).
func createDefaultCurrenciesGorm(db *gorm.DB) error {
	defaultCurrencies := []Currency{
		{Code: "IDR", Name: "Indonesian Rupiah", Symbol: "Rp", IsSystem: true},
		{Code: "USD", Name: "US Dollar", Symbol: "$", IsSystem: true},
		{Code: "EUR", Name: "Euro", Symbol: "€", IsSystem: true},
		{Code: "GBP", Name: "British Pound", Symbol: "£", IsSystem: true},
		{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", IsSystem: true},
		{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", IsSystem: true},
		{Code: "AUD", Name: "Australian Dollar", Symbol: "A$", IsSystem: true},
		{Code: "CHF", Name: "Swiss Franc", Symbol: "CHF", IsSystem: true},
		{Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", IsSystem: true},
		{Code: "INR", Name: "Indian Rupee", Symbol: "₹", IsSystem: true},
		{Code: "BRL", Name: "Brazilian Real", Symbol: "R$", IsSystem: true},
		{Code: "KRW", Name: "South Korean Won", Symbol: "₩", IsSystem: true},
		{Code: "MXN", Name: "Mexican Peso", Symbol: "$", IsSystem: true},
		{Code: "SGD", Name: "Singapore Dollar", Symbol: "S$", IsSystem: true},
		{Code: "HKD", Name: "Hong Kong Dollar", Symbol: "HK$", IsSystem: true},
		{Code: "NZD", Name: "New Zealand Dollar", Symbol: "NZ$", IsSystem: true},
		{Code: "SEK", Name: "Swedish Krona", Symbol: "kr", IsSystem: true},
		{Code: "NOK", Name: "Norwegian Krone", Symbol: "kr", IsSystem: true},
		{Code: "DKK", Name: "Danish Krone", Symbol: "kr", IsSystem: true},
		{Code: "PLN", Name: "Polish Złoty", Symbol: "zł", IsSystem: true},
		{Code: "THB", Name: "Thai Baht", Symbol: "฿", IsSystem: true},
	}

	created := 0
	for _, currency := range defaultCurrencies {
		var existing Currency
		err := db.Where("code = ? AND user_id IS NULL", currency.Code).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Create(&currency).Error; err != nil {
			return err
		}
		created++
	}

	if created > 0 {
		log.Printf("Created %d missing default currencies", created)
	} else {
		log.Println("Default currencies already complete")
	}
	return nil
}

// backfillDefaultWallets creates a default Cash wallet per user and attaches existing rows.
func backfillDefaultWallets(db *gorm.DB) error {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		var walletCount int64
		if err := db.Model(&Wallet{}).Where("user_id = ?", user.ID).Count(&walletCount).Error; err != nil {
			return err
		}

		var defaultWallet Wallet
		if walletCount == 0 {
			currencyID := uint(1)
			var prefs UserPreferences
			if err := db.Where("user_id = ?", user.ID).First(&prefs).Error; err == nil && prefs.PrimaryCurrencyID != 0 {
				currencyID = prefs.PrimaryCurrencyID
			} else {
				var systemCurrency Currency
				if err := db.Where("is_system = ? AND code = ?", true, "IDR").First(&systemCurrency).Error; err == nil {
					currencyID = systemCurrency.ID
				} else if err := db.Where("is_system = ? AND code = ?", true, "USD").First(&systemCurrency).Error; err == nil {
					currencyID = systemCurrency.ID
				} else if err := db.Where("is_system = ?", true).First(&systemCurrency).Error; err == nil {
					currencyID = systemCurrency.ID
				}
			}

			defaultWallet = Wallet{
				UserID:         user.ID,
				Name:           "Cash",
				Type:           "cash",
				CurrencyID:     currencyID,
				OpeningBalance: 0,
				IsDefault:      true,
				IsArchived:     false,
			}
			if err := db.Create(&defaultWallet).Error; err != nil {
				return err
			}
			log.Printf("Created default wallet for public_id=%s", user.PublicID)
		} else {
			if err := db.Where("user_id = ? AND is_default = ?", user.ID, true).First(&defaultWallet).Error; err != nil {
				if err := db.Where("user_id = ? AND is_archived = ?", user.ID, false).First(&defaultWallet).Error; err != nil {
					if err := db.Where("user_id = ?", user.ID).First(&defaultWallet).Error; err != nil {
						return err
					}
				}
			}
		}

		walletID := defaultWallet.ID
		updates := map[string]interface{}{"wallet_id": walletID}
		if err := db.Model(&Expense{}).Where("user_id = ? AND wallet_id IS NULL", user.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := db.Model(&Income{}).Where("user_id = ? AND wallet_id IS NULL", user.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := db.Model(&RecurringTransaction{}).Where("user_id = ? AND wallet_id IS NULL", user.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := db.Model(&PendingTransaction{}).Where("user_id = ? AND wallet_id IS NULL", user.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	return nil
}

// backfillBudgetLimitTypes sets limit_type=fixed for legacy budget rows.
func backfillBudgetLimitTypes(db *gorm.DB) error {
	return db.Model(&Budget{}).
		Where("limit_type = '' OR limit_type IS NULL").
		Update("limit_type", "fixed").Error
}

// backfillFreeSubscriptions ensures every user has a Free subscription row.
func backfillFreeSubscriptions(db *gorm.DB) error {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		var count int64
		if err := db.Model(&Subscription{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		if user.PublicID == "" {
			return fmt.Errorf("user missing public_id")
		}
		sub := Subscription{
			UserID:            user.ID,
			Plan:              "free",
			Status:            "expired",
			DoitCustomerRef:   fmt.Sprintf("user:%s", user.PublicID),
			CancelAtPeriodEnd: false,
		}
		if err := db.Create(&sub).Error; err != nil {
			return err
		}
		log.Printf("Created free subscription for public_id=%s", user.PublicID)
	}

	return nil
}

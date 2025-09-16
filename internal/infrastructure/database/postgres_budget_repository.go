package database

import (
	"context"
	"database/sql"
	"panda-pocket/internal/domain/finance"
	"time"
)

// PostgresBudgetRepository implements the BudgetRepository interface using PostgreSQL
type PostgresBudgetRepository struct {
	db *sql.DB
}

// NewPostgresBudgetRepository creates a new PostgreSQL budget repository
func NewPostgresBudgetRepository(db *sql.DB) *PostgresBudgetRepository {
	return &PostgresBudgetRepository{db: db}
}

// Save saves a budget to the database
func (r *PostgresBudgetRepository) Save(ctx context.Context, budget *finance.Budget) error {
	if budget.ID().Value() == 0 {
		// Insert new budget
		query := `INSERT INTO budgets (user_id, category_id, amount, period, start_date, end_date, created_at) 
				  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

		var id int
		err := r.db.QueryRowContext(ctx, query,
			budget.UserID().Value(),
			budget.CategoryID().Value(),
			budget.Amount().Amount(),
			string(budget.Period()),
			budget.StartDate(),
			budget.EndDate(),
			budget.CreatedAt(),
		).Scan(&id)

		if err != nil {
			return err
		}

		// Note: In a real implementation, you'd want to handle the ID assignment properly
		_ = id
	} else {
		// Update existing budget
		query := `UPDATE budgets SET amount = $1, period = $2, start_date = $3, end_date = $4 
				  WHERE id = $5`
		_, err := r.db.ExecContext(ctx, query,
			budget.Amount().Amount(),
			string(budget.Period()),
			budget.StartDate(),
			budget.EndDate(),
			budget.ID().Value(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindByID finds a budget by ID
func (r *PostgresBudgetRepository) FindByID(ctx context.Context, id finance.BudgetID) (*finance.Budget, error) {
	var userID, categoryID int
	var amount float64
	var period string
	var startDate, endDate, createdAt time.Time

	err := r.db.QueryRowContext(ctx,
		"SELECT user_id, category_id, amount, period, start_date, end_date, created_at FROM budgets WHERE id = $1",
		id.Value(),
	).Scan(&userID, &categoryID, &amount, &period, &startDate, &endDate, &createdAt)

	if err != nil {
		return nil, err
	}

	// Create money object (assuming default currency for now)
	money, err := finance.NewMoney(amount, finance.NewCurrencyID(1)) // Default currency ID
	if err != nil {
		return nil, err
	}

	budget, err := finance.NewBudget(
		id,
		finance.NewUserID(userID),
		finance.NewCategoryID(categoryID),
		money,
		finance.BudgetPeriod(period),
		startDate,
	)

	if err != nil {
		return nil, err
	}

	return budget, nil
}

// FindByUserID finds all budgets for a user
func (r *PostgresBudgetRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Budget, error) {
	var budgets []*finance.Budget

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, category_id, amount, period, start_date, end_date, created_at FROM budgets WHERE user_id = $1 ORDER BY created_at DESC",
		userID.Value(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, categoryID int
		var amount float64
		var period string
		var startDate, endDate, createdAt time.Time

		err := rows.Scan(&id, &categoryID, &amount, &period, &startDate, &endDate, &createdAt)
		if err != nil {
			continue
		}

		// Create money object (assuming default currency for now)
		money, err := finance.NewMoney(amount, finance.NewCurrencyID(1)) // Default currency ID
		if err != nil {
			continue
		}

		budget, err := finance.NewBudget(
			finance.NewBudgetID(id),
			userID,
			finance.NewCategoryID(categoryID),
			money,
			finance.BudgetPeriod(period),
			startDate,
		)

		if err != nil {
			continue
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// FindByUserIDAndCategory finds budgets for a user by category
func (r *PostgresBudgetRepository) FindByUserIDAndCategory(ctx context.Context, userID finance.UserID, categoryID finance.CategoryID) ([]*finance.Budget, error) {
	var budgets []*finance.Budget

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, amount, period, start_date, end_date, created_at FROM budgets WHERE user_id = $1 AND category_id = $2 ORDER BY created_at DESC",
		userID.Value(), categoryID.Value(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var amount float64
		var period string
		var startDate, endDate, createdAt time.Time

		err := rows.Scan(&id, &amount, &period, &startDate, &endDate, &createdAt)
		if err != nil {
			continue
		}

		// Create money object (assuming default currency for now)
		money, err := finance.NewMoney(amount, finance.NewCurrencyID(1)) // Default currency ID
		if err != nil {
			continue
		}

		budget, err := finance.NewBudget(
			finance.NewBudgetID(id),
			userID,
			categoryID,
			money,
			finance.BudgetPeriod(period),
			startDate,
		)

		if err != nil {
			continue
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// FindActiveByUserID finds active budgets for a user
func (r *PostgresBudgetRepository) FindActiveByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Budget, error) {
	var budgets []*finance.Budget
	now := time.Now()

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, category_id, amount, period, start_date, end_date, created_at FROM budgets WHERE user_id = $1 AND start_date <= $2 AND end_date >= $2 ORDER BY created_at DESC",
		userID.Value(), now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, categoryID int
		var amount float64
		var period string
		var startDate, endDate, createdAt time.Time

		err := rows.Scan(&id, &categoryID, &amount, &period, &startDate, &endDate, &createdAt)
		if err != nil {
			continue
		}

		// Create money object (assuming default currency for now)
		money, err := finance.NewMoney(amount, finance.NewCurrencyID(1)) // Default currency ID
		if err != nil {
			continue
		}

		budget, err := finance.NewBudget(
			finance.NewBudgetID(id),
			userID,
			finance.NewCategoryID(categoryID),
			money,
			finance.BudgetPeriod(period),
			startDate,
		)

		if err != nil {
			continue
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// Delete deletes a budget by ID
func (r *PostgresBudgetRepository) Delete(ctx context.Context, id finance.BudgetID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM budgets WHERE id = $1", id.Value())
	return err
}

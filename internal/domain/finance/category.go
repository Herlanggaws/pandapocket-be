package finance

import (
	"errors"
	"time"
)

// CategoryType represents the type of category
type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
)

// Category represents a transaction category
type Category struct {
	id           CategoryID
	userID       *UserID // nil for default categories
	name         string
	color        string
	isDefault    bool
	categoryType CategoryType
	createdAt    time.Time
}

// NewCategory creates a new category
func NewCategory(
	id CategoryID,
	userID *UserID,
	name string,
	color string,
	isDefault bool,
	categoryType CategoryType,
) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	
	if color == "" {
		color = "#3B82F6" // Default color
	}
	
	return &Category{
		id:           id,
		userID:       userID,
		name:         name,
		color:        color,
		isDefault:    isDefault,
		categoryType: categoryType,
		createdAt:    time.Now(),
	}, nil
}

// Getters
func (c *Category) ID() CategoryID {
	return c.id
}

func (c *Category) UserID() *UserID {
	return c.userID
}

func (c *Category) Name() string {
	return c.name
}

func (c *Category) Color() string {
	return c.color
}

func (c *Category) IsDefault() bool {
	return c.isDefault
}

func (c *Category) Type() CategoryType {
	return c.categoryType
}

func (c *Category) CreatedAt() time.Time {
	return c.createdAt
}

// UpdateName updates the category name
func (c *Category) UpdateName(name string) error {
	if name == "" {
		return errors.New("category name cannot be empty")
	}
	c.name = name
	return nil
}

// UpdateColor updates the category color
func (c *Category) UpdateColor(color string) {
	if color == "" {
		color = "#3B82F6"
	}
	c.color = color
}

// CanBeDeleted checks if the category can be deleted
func (c *Category) CanBeDeleted() bool {
	return !c.isDefault
}

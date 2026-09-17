package finance

import (
	"context"
	"time"
)

type HealthScoreSnapshot struct {
	id               int
	userID           UserID
	yearMonth        string
	score            int
	budgetAdherence  float64
	cashflow         float64
	coverage         float64
	computedAt       time.Time
}

func NewHealthScoreSnapshot(
	userID UserID,
	yearMonth string,
	score int,
	budgetAdherence, cashflow, coverage float64,
) *HealthScoreSnapshot {
	return &HealthScoreSnapshot{
		userID:          userID,
		yearMonth:       yearMonth,
		score:           score,
		budgetAdherence: budgetAdherence,
		cashflow:        cashflow,
		coverage:        coverage,
		computedAt:      time.Now(),
	}
}

func ReconstituteHealthScoreSnapshot(
	id int,
	userID UserID,
	yearMonth string,
	score int,
	budgetAdherence, cashflow, coverage float64,
	computedAt time.Time,
) *HealthScoreSnapshot {
	return &HealthScoreSnapshot{
		id:              id,
		userID:          userID,
		yearMonth:       yearMonth,
		score:           score,
		budgetAdherence: budgetAdherence,
		cashflow:        cashflow,
		coverage:        coverage,
		computedAt:      computedAt,
	}
}

func (s *HealthScoreSnapshot) ID() int                    { return s.id }
func (s *HealthScoreSnapshot) UserID() UserID             { return s.userID }
func (s *HealthScoreSnapshot) YearMonth() string          { return s.yearMonth }
func (s *HealthScoreSnapshot) Score() int                 { return s.score }
func (s *HealthScoreSnapshot) BudgetAdherence() float64   { return s.budgetAdherence }
func (s *HealthScoreSnapshot) Cashflow() float64          { return s.cashflow }
func (s *HealthScoreSnapshot) Coverage() float64          { return s.coverage }
func (s *HealthScoreSnapshot) ComputedAt() time.Time      { return s.computedAt }
func (s *HealthScoreSnapshot) AssignID(id int)            { s.id = id }

type HealthScoreSnapshotRepository interface {
	Upsert(ctx context.Context, snapshot *HealthScoreSnapshot) error
	FindByUserID(ctx context.Context, userID UserID, limit int) ([]*HealthScoreSnapshot, error)
}

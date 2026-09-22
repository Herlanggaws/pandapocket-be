package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormGoalRepository struct {
	db *gorm.DB
}

func NewGormGoalRepository(db *gorm.DB) *GormGoalRepository {
	return &GormGoalRepository{db: db}
}

func (r *GormGoalRepository) toDomain(model FinancialGoal) *finance.FinancialGoal {
	var walletID *finance.WalletID
	if model.WalletID != nil {
		id := finance.NewWalletID(int(*model.WalletID))
		walletID = &id
	}
	return finance.ReconstituteFinancialGoal(
		finance.NewGoalID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		model.Name,
		model.TargetAmount,
		finance.NewCurrencyID(int(model.CurrencyID)),
		model.CurrentAmount,
		model.TargetDate,
		finance.GoalStatus(model.Status),
		walletID,
		model.CreatedAt,
	)
}

func (r *GormGoalRepository) Save(ctx context.Context, goal *finance.FinancialGoal) error {
	model := &FinancialGoal{
		UserID:        uint(goal.UserID().Value()),
		Name:          goal.Name(),
		TargetAmount:  goal.TargetAmount(),
		CurrencyID:    uint(goal.CurrencyID().Value()),
		CurrentAmount: goal.CurrentAmount(),
		TargetDate:    goal.TargetDate(),
		Status:        string(goal.Status()),
	}
	if goal.WalletID() != nil {
		id := uint(goal.WalletID().Value())
		model.WalletID = &id
	}
	if goal.ID().Value() != 0 {
		model.ID = uint(goal.ID().Value())
		return r.db.WithContext(ctx).Model(&FinancialGoal{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"name":           model.Name,
			"target_amount":  model.TargetAmount,
			"currency_id":    model.CurrencyID,
			"current_amount": model.CurrentAmount,
			"target_date":    model.TargetDate,
			"status":         model.Status,
			"wallet_id":      model.WalletID,
			"updated_at":     time.Now(),
		}).Error
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	goal.AssignID(finance.NewGoalID(int(model.ID)))
	return nil
}

func (r *GormGoalRepository) FindByID(ctx context.Context, id finance.GoalID) (*finance.FinancialGoal, error) {
	var model FinancialGoal
	if err := r.db.WithContext(ctx).First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("goal not found")
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormGoalRepository) FindByUserID(ctx context.Context, userID finance.UserID, includeArchived bool) ([]*finance.FinancialGoal, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value())
	if !includeArchived {
		query = query.Where("status <> ?", "archived")
	}
	var models []FinancialGoal
	if err := query.Order("target_date ASC, id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.FinancialGoal, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

func (r *GormGoalRepository) FindByWalletID(ctx context.Context, walletID finance.WalletID) ([]*finance.FinancialGoal, error) {
	var models []FinancialGoal
	if err := r.db.WithContext(ctx).
		Where("wallet_id = ?", walletID.Value()).
		Order("id ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.FinancialGoal, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

func (r *GormGoalRepository) FindActiveForDeadlineDates(ctx context.Context, dates []time.Time) ([]*finance.FinancialGoal, error) {
	if len(dates) == 0 {
		return []*finance.FinancialGoal{}, nil
	}
	dayStrings := make([]string, 0, len(dates))
	for _, d := range dates {
		dayStrings = append(dayStrings, d.Format("2006-01-02"))
	}
	var models []FinancialGoal
	if err := r.db.WithContext(ctx).
		Where("status = ? AND target_date IN ?", "active", dayStrings).
		Order("user_id ASC, id ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.FinancialGoal, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

func (r *GormGoalRepository) Delete(ctx context.Context, id finance.GoalID, userID finance.UserID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id.Value(), userID.Value()).Delete(&FinancialGoal{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("goal not found")
	}
	return nil
}

type GormAssetRepository struct {
	db *gorm.DB
}

func NewGormAssetRepository(db *gorm.DB) *GormAssetRepository {
	return &GormAssetRepository{db: db}
}

func (r *GormAssetRepository) toDomain(model Asset) *finance.Asset {
	return finance.ReconstituteAsset(
		finance.NewAssetID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		model.Name,
		finance.AssetType(model.Type),
		finance.NewCurrencyID(int(model.CurrencyID)),
		model.CurrentValue,
		model.Notes,
		model.IsArchived,
		model.AsOfDate,
		model.CreatedAt,
	)
}

func (r *GormAssetRepository) Save(ctx context.Context, asset *finance.Asset) error {
	model := &Asset{
		UserID:       uint(asset.UserID().Value()),
		Name:         asset.Name(),
		Type:         string(asset.Type()),
		CurrencyID:   uint(asset.CurrencyID().Value()),
		CurrentValue: asset.CurrentValue(),
		Notes:        asset.Notes(),
		IsArchived:   asset.IsArchived(),
		AsOfDate:     asset.AsOfDate(),
	}
	if asset.ID().Value() != 0 {
		model.ID = uint(asset.ID().Value())
		return r.db.WithContext(ctx).Model(&Asset{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"name":          model.Name,
			"type":          model.Type,
			"current_value": model.CurrentValue,
			"notes":         model.Notes,
			"is_archived":   model.IsArchived,
			"as_of_date":    model.AsOfDate,
			"updated_at":    time.Now(),
		}).Error
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	asset.AssignID(finance.NewAssetID(int(model.ID)))
	return nil
}

func (r *GormAssetRepository) FindByID(ctx context.Context, id finance.AssetID) (*finance.Asset, error) {
	var model Asset
	if err := r.db.WithContext(ctx).First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("asset not found")
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormAssetRepository) FindByUserID(ctx context.Context, userID finance.UserID, includeArchived bool) ([]*finance.Asset, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value())
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}
	var models []Asset
	if err := query.Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.Asset, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

type GormLiabilityRepository struct {
	db *gorm.DB
}

func NewGormLiabilityRepository(db *gorm.DB) *GormLiabilityRepository {
	return &GormLiabilityRepository{db: db}
}

func (r *GormLiabilityRepository) toDomain(model Liability) *finance.Liability {
	return finance.ReconstituteLiability(
		finance.NewLiabilityID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		model.Name,
		finance.LiabilityType(model.Type),
		finance.NewCurrencyID(int(model.CurrencyID)),
		model.CurrentBalance,
		model.Notes,
		model.IsArchived,
		model.AsOfDate,
		model.CreatedAt,
		finance.LiabilityDebtDetails{
			OriginalPrincipal: model.OriginalPrincipal,
			InterestRateAPR:   model.InterestRateAPR,
			MinimumPayment:    model.MinimumPayment,
			NextDueDate:       model.NextDueDate,
		},
	)
}

func (r *GormLiabilityRepository) Save(ctx context.Context, liability *finance.Liability) error {
	model := &Liability{
		UserID:            uint(liability.UserID().Value()),
		Name:              liability.Name(),
		Type:              string(liability.Type()),
		CurrencyID:        uint(liability.CurrencyID().Value()),
		CurrentBalance:    liability.CurrentBalance(),
		OriginalPrincipal: liability.OriginalPrincipal(),
		InterestRateAPR:   liability.InterestRateAPR(),
		MinimumPayment:    liability.MinimumPayment(),
		NextDueDate:       liability.NextDueDate(),
		Notes:             liability.Notes(),
		IsArchived:        liability.IsArchived(),
		AsOfDate:          liability.AsOfDate(),
	}
	if liability.ID().Value() != 0 {
		model.ID = uint(liability.ID().Value())
		return r.db.WithContext(ctx).Model(&Liability{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"name":               model.Name,
			"type":               model.Type,
			"current_balance":    model.CurrentBalance,
			"original_principal": model.OriginalPrincipal,
			"interest_rate_apr":  model.InterestRateAPR,
			"minimum_payment":    model.MinimumPayment,
			"next_due_date":      model.NextDueDate,
			"notes":              model.Notes,
			"is_archived":        model.IsArchived,
			"as_of_date":         model.AsOfDate,
			"updated_at":         time.Now(),
		}).Error
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	liability.AssignID(finance.NewLiabilityID(int(model.ID)))
	return nil
}

func (r *GormLiabilityRepository) FindByID(ctx context.Context, id finance.LiabilityID) (*finance.Liability, error) {
	var model Liability
	if err := r.db.WithContext(ctx).First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("liability not found")
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormLiabilityRepository) FindByUserID(ctx context.Context, userID finance.UserID, includeArchived bool) ([]*finance.Liability, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value())
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}
	var models []Liability
	if err := query.Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.Liability, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

type GormLiabilityPaymentRepository struct {
	db *gorm.DB
}

func NewGormLiabilityPaymentRepository(db *gorm.DB) *GormLiabilityPaymentRepository {
	return &GormLiabilityPaymentRepository{db: db}
}

func (r *GormLiabilityPaymentRepository) Save(ctx context.Context, payment *finance.LiabilityPayment) error {
	model := &LiabilityPayment{
		LiabilityID: uint(payment.LiabilityID().Value()),
		UserID:      uint(payment.UserID().Value()),
		Amount:      payment.Amount(),
		PaidAt:      payment.PaidAt(),
		Note:        payment.Note(),
	}
	if payment.ExpenseID() != nil {
		id := uint(*payment.ExpenseID())
		model.ExpenseID = &id
	}
	if payment.ID().Value() != 0 {
		model.ID = uint(payment.ID().Value())
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	payment.AssignID(finance.NewLiabilityPaymentID(int(model.ID)))
	return nil
}

func (r *GormLiabilityPaymentRepository) FindByLiabilityID(ctx context.Context, liabilityID finance.LiabilityID) ([]*finance.LiabilityPayment, error) {
	var models []LiabilityPayment
	if err := r.db.WithContext(ctx).
		Where("liability_id = ?", liabilityID.Value()).
		Order("paid_at DESC, id DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.LiabilityPayment, 0, len(models))
	for _, m := range models {
		result = append(result, toDomainLiabilityPayment(m))
	}
	return result, nil
}

func (r *GormLiabilityPaymentRepository) FindByExpenseID(ctx context.Context, expenseID int) (*finance.LiabilityPayment, error) {
	var model LiabilityPayment
	err := r.db.WithContext(ctx).Where("expense_id = ?", expenseID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainLiabilityPayment(model), nil
}

func (r *GormLiabilityPaymentRepository) Delete(ctx context.Context, id finance.LiabilityPaymentID) error {
	return r.db.WithContext(ctx).Delete(&LiabilityPayment{}, id.Value()).Error
}

func toDomainLiabilityPayment(m LiabilityPayment) *finance.LiabilityPayment {
	var expenseID *int
	if m.ExpenseID != nil {
		v := int(*m.ExpenseID)
		expenseID = &v
	}
	return finance.ReconstituteLiabilityPayment(
		finance.NewLiabilityPaymentID(int(m.ID)),
		finance.NewLiabilityID(int(m.LiabilityID)),
		finance.NewUserID(int(m.UserID)),
		m.Amount,
		m.PaidAt,
		expenseID,
		m.Note,
		m.CreatedAt,
	)
}

type GormGoalContributionRepository struct {
	db *gorm.DB
}

func NewGormGoalContributionRepository(db *gorm.DB) *GormGoalContributionRepository {
	return &GormGoalContributionRepository{db: db}
}

func toDomainGoalContribution(m GoalContribution) *finance.GoalContribution {
	var expenseID, incomeID, transferID *int
	if m.ExpenseID != nil {
		v := int(*m.ExpenseID)
		expenseID = &v
	}
	if m.IncomeID != nil {
		v := int(*m.IncomeID)
		incomeID = &v
	}
	if m.TransferID != nil {
		v := int(*m.TransferID)
		transferID = &v
	}
	return finance.ReconstituteGoalContribution(
		finance.NewGoalContributionID(int(m.ID)),
		finance.NewGoalID(int(m.GoalID)),
		finance.NewUserID(int(m.UserID)),
		m.Amount,
		m.ContributedAt,
		expenseID,
		incomeID,
		transferID,
		m.Note,
		m.CreatedAt,
	)
}

func (r *GormGoalContributionRepository) Save(ctx context.Context, contribution *finance.GoalContribution) error {
	model := &GoalContribution{
		GoalID:        uint(contribution.GoalID().Value()),
		UserID:        uint(contribution.UserID().Value()),
		Amount:        contribution.Amount(),
		ContributedAt: contribution.ContributedAt(),
		Note:          contribution.Note(),
	}
	if contribution.ExpenseID() != nil {
		id := uint(*contribution.ExpenseID())
		model.ExpenseID = &id
	}
	if contribution.IncomeID() != nil {
		id := uint(*contribution.IncomeID())
		model.IncomeID = &id
	}
	if contribution.TransferID() != nil {
		id := uint(*contribution.TransferID())
		model.TransferID = &id
	}
	if contribution.ID().Value() != 0 {
		model.ID = uint(contribution.ID().Value())
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	contribution.AssignID(finance.NewGoalContributionID(int(model.ID)))
	return nil
}

func (r *GormGoalContributionRepository) FindByGoalID(ctx context.Context, goalID finance.GoalID) ([]*finance.GoalContribution, error) {
	var models []GoalContribution
	if err := r.db.WithContext(ctx).
		Where("goal_id = ?", goalID.Value()).
		Order("contributed_at DESC, id DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.GoalContribution, 0, len(models))
	for _, m := range models {
		result = append(result, toDomainGoalContribution(m))
	}
	return result, nil
}

func (r *GormGoalContributionRepository) FindByExpenseID(ctx context.Context, expenseID int) (*finance.GoalContribution, error) {
	var model GoalContribution
	err := r.db.WithContext(ctx).Where("expense_id = ?", expenseID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainGoalContribution(model), nil
}

func (r *GormGoalContributionRepository) FindByIncomeID(ctx context.Context, incomeID int) (*finance.GoalContribution, error) {
	var model GoalContribution
	err := r.db.WithContext(ctx).Where("income_id = ?", incomeID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainGoalContribution(model), nil
}

func (r *GormGoalContributionRepository) FindByTransferID(ctx context.Context, transferID int) (*finance.GoalContribution, error) {
	var model GoalContribution
	err := r.db.WithContext(ctx).Where("transfer_id = ?", transferID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainGoalContribution(model), nil
}

func (r *GormGoalContributionRepository) Delete(ctx context.Context, id finance.GoalContributionID) error {
	return r.db.WithContext(ctx).Delete(&GoalContribution{}, id.Value()).Error
}

type GormHealthScoreSnapshotRepository struct {
	db *gorm.DB
}

func NewGormHealthScoreSnapshotRepository(db *gorm.DB) *GormHealthScoreSnapshotRepository {
	return &GormHealthScoreSnapshotRepository{db: db}
}

func (r *GormHealthScoreSnapshotRepository) Upsert(ctx context.Context, snapshot *finance.HealthScoreSnapshot) error {
	model := HealthScoreSnapshot{
		UserID:          uint(snapshot.UserID().Value()),
		YearMonth:       snapshot.YearMonth(),
		Score:           snapshot.Score(),
		BudgetAdherence: snapshot.BudgetAdherence(),
		Cashflow:        snapshot.Cashflow(),
		Coverage:        snapshot.Coverage(),
		ComputedAt:      snapshot.ComputedAt(),
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "year_month"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"score", "budget_adherence", "cashflow", "coverage", "computed_at", "updated_at",
		}),
	}).Create(&model).Error
	if err != nil {
		return err
	}
	if snapshot.ID() == 0 {
		snapshot.AssignID(int(model.ID))
	}
	return nil
}

func (r *GormHealthScoreSnapshotRepository) FindByUserID(ctx context.Context, userID finance.UserID, limit int) ([]*finance.HealthScoreSnapshot, error) {
	if limit <= 0 {
		limit = 12
	}
	var models []HealthScoreSnapshot
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID.Value()).
		Order("year_month DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.HealthScoreSnapshot, 0, len(models))
	for _, m := range models {
		result = append(result, finance.ReconstituteHealthScoreSnapshot(
			int(m.ID),
			finance.NewUserID(int(m.UserID)),
			m.YearMonth,
			m.Score,
			m.BudgetAdherence,
			m.Cashflow,
			m.Coverage,
			m.ComputedAt,
		))
	}
	return result, nil
}

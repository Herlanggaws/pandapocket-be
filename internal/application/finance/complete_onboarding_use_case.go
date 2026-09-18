package finance

import (
	"context"
	"encoding/json"
	"errors"
	"panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
	"time"

	"gorm.io/gorm"
)

type CompleteOnboardingRequest struct {
	PrimaryCurrencyID *int     `json:"primary_currency_id"`
	Goal              string   `json:"goal"`
	Topics            []string `json:"topics"`
	MonthlyIncome     float64  `json:"monthly_income" binding:"required,gte=0"`
	MonthlyExpense    float64  `json:"monthly_expense" binding:"required,gte=0"`
	DebtBalance       *float64 `json:"debt_balance"`
	DebtName          string   `json:"debt_name"`
	DebtType          string   `json:"debt_type"`
}

type CompleteOnboardingResponse struct {
	Onboarding  map[string]interface{} `json:"onboarding"`
	HealthScore *HealthScoreResponse   `json:"health_score,omitempty"`
	Message     string                 `json:"message"`
}

type onboardingPreferencesRepo interface {
	FindByUserID(ctx context.Context, userID domainIdentity.UserID) (*domainIdentity.UserPreferences, error)
	Save(ctx context.Context, prefs *domainIdentity.UserPreferences) error
}

type CompleteOnboardingUseCase struct {
	prefsRepo                onboardingPreferencesRepo
	walletService            *finance.WalletService
	currencyService          *finance.CurrencyService
	categoryService          *finance.CategoryService
	createTransactionUseCase *CreateTransactionUseCase
	createBudgetUseCase      *CreateBudgetUseCase
	liabilityService         *finance.LiabilityService
	getHealthScoreUseCase    *GetHealthScoreUseCase
}

func NewCompleteOnboardingUseCase(
	prefsRepo onboardingPreferencesRepo,
	walletService *finance.WalletService,
	currencyService *finance.CurrencyService,
	categoryService *finance.CategoryService,
	createTransactionUseCase *CreateTransactionUseCase,
	createBudgetUseCase *CreateBudgetUseCase,
	liabilityService *finance.LiabilityService,
	getHealthScoreUseCase *GetHealthScoreUseCase,
) *CompleteOnboardingUseCase {
	return &CompleteOnboardingUseCase{
		prefsRepo:                prefsRepo,
		walletService:            walletService,
		currencyService:          currencyService,
		categoryService:          categoryService,
		createTransactionUseCase: createTransactionUseCase,
		createBudgetUseCase:      createBudgetUseCase,
		liabilityService:         liabilityService,
		getHealthScoreUseCase:    getHealthScoreUseCase,
	}
}

func (uc *CompleteOnboardingUseCase) findCategoryID(ctx context.Context, userID int, categoryType finance.CategoryType, preferredNames ...string) (int, error) {
	categories, err := uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), categoryType)
	if err != nil {
		return 0, err
	}
	for _, name := range preferredNames {
		for _, category := range categories {
			if category.Name() == name {
				return category.ID().Value(), nil
			}
		}
	}
	if len(categories) == 0 {
		return 0, errors.New("no categories available")
	}
	return categories[0].ID().Value(), nil
}

func (uc *CompleteOnboardingUseCase) Execute(ctx context.Context, userID int, req CompleteOnboardingRequest) (*CompleteOnboardingResponse, error) {
	user := domainIdentity.NewUserID(userID)

	prefs, err := uc.prefsRepo.FindByUserID(ctx, user)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		currency, cerr := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
		if cerr != nil {
			return nil, errors.New("unable to load preferences")
		}
		prefs = domainIdentity.NewUserPreferences(user, currency.ID().Value(), true, true, true, json.RawMessage("{}"))
	}

	if req.PrimaryCurrencyID != nil && *req.PrimaryCurrencyID > 0 {
		prefs.SetPrimaryCurrencyID(*req.PrimaryCurrencyID)
	}

	onboardingMap := map[string]interface{}{}
	_ = json.Unmarshal(prefs.Onboarding(), &onboardingMap)
	onboardingMap["onboarding_completed"] = true
	if req.Goal != "" {
		onboardingMap["goal"] = req.Goal
	}
	if req.Topics != nil {
		onboardingMap["topics"] = req.Topics
	}
	onboardingMap["monthly_income"] = req.MonthlyIncome
	onboardingMap["monthly_expense"] = req.MonthlyExpense
	if req.DebtBalance != nil {
		onboardingMap["debt_balance"] = *req.DebtBalance
	}
	merged, err := json.Marshal(onboardingMap)
	if err != nil {
		return nil, err
	}
	prefs.SetOnboarding(merged)
	if err := uc.prefsRepo.Save(ctx, prefs); err != nil {
		return nil, err
	}

	primaryCurrencyID := prefs.PrimaryCurrencyID()
	_, err = uc.walletService.EnsureDefaultWallet(ctx, finance.NewUserID(userID), finance.NewCurrencyID(primaryCurrencyID))
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location()).Format("2006-01-02")

	if req.MonthlyIncome > 0 {
		incomeCategoryID, catErr := uc.findCategoryID(ctx, userID, finance.CategoryTypeIncome, "Salary", "Other")
		if catErr != nil {
			return nil, catErr
		}
		if _, err := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
			CategoryID:  incomeCategoryID,
			Amount:      req.MonthlyIncome,
			Description: "Onboarding monthly income",
			Date:        today,
			Type:        "income",
		}); err != nil {
			return nil, err
		}
	}

	if req.MonthlyExpense > 0 {
		expenseCategoryID, catErr := uc.findCategoryID(ctx, userID, finance.CategoryTypeExpense, "Other", "Bills", "Food")
		if catErr != nil {
			return nil, catErr
		}
		if _, err := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
			CategoryID:  expenseCategoryID,
			Amount:      req.MonthlyExpense,
			Description: "Onboarding monthly expense estimate",
			Date:        today,
			Type:        "expense",
		}); err != nil {
			return nil, err
		}

		if _, err := uc.createBudgetUseCase.Execute(ctx, userID, CreateBudgetRequest{
			CategoryID: expenseCategoryID,
			Amount:     req.MonthlyExpense,
			LimitType:  "fixed",
			Period:     "monthly",
			StartDate:  monthStart,
		}); err != nil {
			return nil, err
		}
	}

	if req.Goal == "debt" && req.DebtBalance != nil && *req.DebtBalance > 0 {
		debtType := req.DebtType
		if debtType == "" {
			debtType = "other"
		}
		liabilityType, parseErr := finance.ParseLiabilityType(debtType)
		if parseErr != nil {
			liabilityType = finance.LiabilityTypeOther
		}
		name := req.DebtName
		if name == "" {
			name = "Onboarding debt"
		}
		principal := *req.DebtBalance
		if _, err := uc.liabilityService.Create(
			ctx,
			finance.NewUserID(userID),
			name,
			liabilityType,
			finance.NewCurrencyID(primaryCurrencyID),
			*req.DebtBalance,
			"Seeded from onboarding",
			nil,
			finance.LiabilityDebtDetails{OriginalPrincipal: &principal},
		); err != nil {
			return nil, err
		}
	}

	var health *HealthScoreResponse
	if uc.getHealthScoreUseCase != nil {
		health, _ = uc.getHealthScoreUseCase.Execute(ctx, userID)
	}

	return &CompleteOnboardingResponse{
		Onboarding:  onboardingMap,
		HealthScore: health,
		Message:     "Onboarding completed",
	}, nil
}

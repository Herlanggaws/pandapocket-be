package finance

import (
	"errors"
	"math"
	"time"
)

type AssetType string

const (
	AssetTypeProperty   AssetType = "property"
	AssetTypeVehicle    AssetType = "vehicle"
	AssetTypeInvestment AssetType = "investment"
	AssetTypeOther      AssetType = "other"
)

func ParseAssetType(value string) (AssetType, error) {
	switch AssetType(value) {
	case AssetTypeProperty, AssetTypeVehicle, AssetTypeInvestment, AssetTypeOther:
		return AssetType(value), nil
	default:
		return "", errors.New("invalid asset type")
	}
}

type AssetID struct{ value int }

func NewAssetID(id int) AssetID { return AssetID{value: id} }
func (a AssetID) Value() int    { return a.value }

type Asset struct {
	id           AssetID
	userID       UserID
	name         string
	assetType    AssetType
	currencyID   CurrencyID
	currentValue float64
	notes        string
	isArchived   bool
	asOfDate     *time.Time
	createdAt    time.Time
}

func NewAsset(
	userID UserID,
	name string,
	assetType AssetType,
	currencyID CurrencyID,
	currentValue float64,
	notes string,
	asOfDate *time.Time,
) (*Asset, error) {
	if name == "" {
		return nil, errors.New("asset name cannot be empty")
	}
	if _, err := ParseAssetType(string(assetType)); err != nil {
		return nil, err
	}
	if currentValue < 0 {
		return nil, errors.New("current value cannot be negative")
	}
	return &Asset{
		userID:       userID,
		name:         name,
		assetType:    assetType,
		currencyID:   currencyID,
		currentValue: currentValue,
		notes:        notes,
		isArchived:   false,
		asOfDate:     asOfDate,
		createdAt:    time.Now(),
	}, nil
}

func ReconstituteAsset(
	id AssetID,
	userID UserID,
	name string,
	assetType AssetType,
	currencyID CurrencyID,
	currentValue float64,
	notes string,
	isArchived bool,
	asOfDate *time.Time,
	createdAt time.Time,
) *Asset {
	return &Asset{
		id:           id,
		userID:       userID,
		name:         name,
		assetType:    assetType,
		currencyID:   currencyID,
		currentValue: currentValue,
		notes:        notes,
		isArchived:   isArchived,
		asOfDate:     asOfDate,
		createdAt:    createdAt,
	}
}

func (a *Asset) AssignID(id AssetID)           { a.id = id }
func (a *Asset) ID() AssetID                   { return a.id }
func (a *Asset) UserID() UserID                { return a.userID }
func (a *Asset) Name() string                  { return a.name }
func (a *Asset) Type() AssetType               { return a.assetType }
func (a *Asset) CurrencyID() CurrencyID        { return a.currencyID }
func (a *Asset) CurrentValue() float64         { return a.currentValue }
func (a *Asset) Notes() string                 { return a.notes }
func (a *Asset) IsArchived() bool              { return a.isArchived }
func (a *Asset) AsOfDate() *time.Time          { return a.asOfDate }
func (a *Asset) CreatedAt() time.Time          { return a.createdAt }

func (a *Asset) Update(name string, assetType AssetType, currentValue float64, notes string, asOfDate *time.Time) error {
	if name == "" {
		return errors.New("asset name cannot be empty")
	}
	if _, err := ParseAssetType(string(assetType)); err != nil {
		return err
	}
	if currentValue < 0 {
		return errors.New("current value cannot be negative")
	}
	a.name = name
	a.assetType = assetType
	a.currentValue = currentValue
	a.notes = notes
	a.asOfDate = asOfDate
	return nil
}

func (a *Asset) Archive()   { a.isArchived = true }
func (a *Asset) Unarchive() { a.isArchived = false }

type LiabilityType string

const (
	LiabilityTypeLoan       LiabilityType = "loan"
	LiabilityTypeCreditCard LiabilityType = "credit_card"
	LiabilityTypeMortgage   LiabilityType = "mortgage"
	LiabilityTypeOther      LiabilityType = "other"
)

func ParseLiabilityType(value string) (LiabilityType, error) {
	switch LiabilityType(value) {
	case LiabilityTypeLoan, LiabilityTypeCreditCard, LiabilityTypeMortgage, LiabilityTypeOther:
		return LiabilityType(value), nil
	default:
		return "", errors.New("invalid liability type")
	}
}

type LiabilityID struct{ value int }

func NewLiabilityID(id int) LiabilityID { return LiabilityID{value: id} }
func (l LiabilityID) Value() int        { return l.value }

type Liability struct {
	id                 LiabilityID
	userID             UserID
	name               string
	liabilityType      LiabilityType
	currencyID         CurrencyID
	currentBalance     float64
	originalPrincipal  *float64
	interestRateAPR    *float64
	minimumPayment     float64
	nextDueDate        *time.Time
	notes              string
	isArchived         bool
	asOfDate           *time.Time
	createdAt          time.Time
}

type LiabilityDebtDetails struct {
	OriginalPrincipal *float64
	InterestRateAPR   *float64
	MinimumPayment    float64
	NextDueDate       *time.Time
}

func NewLiability(
	userID UserID,
	name string,
	liabilityType LiabilityType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
	debt LiabilityDebtDetails,
) (*Liability, error) {
	if name == "" {
		return nil, errors.New("liability name cannot be empty")
	}
	if _, err := ParseLiabilityType(string(liabilityType)); err != nil {
		return nil, err
	}
	if currentBalance < 0 {
		return nil, errors.New("current balance cannot be negative")
	}
	if debt.MinimumPayment < 0 {
		return nil, errors.New("minimum payment cannot be negative")
	}
	if debt.OriginalPrincipal != nil && *debt.OriginalPrincipal < 0 {
		return nil, errors.New("original principal cannot be negative")
	}
	if debt.InterestRateAPR != nil && *debt.InterestRateAPR < 0 {
		return nil, errors.New("interest rate cannot be negative")
	}
	return &Liability{
		userID:            userID,
		name:              name,
		liabilityType:     liabilityType,
		currencyID:        currencyID,
		currentBalance:    currentBalance,
		originalPrincipal: debt.OriginalPrincipal,
		interestRateAPR:   debt.InterestRateAPR,
		minimumPayment:    debt.MinimumPayment,
		nextDueDate:       debt.NextDueDate,
		notes:             notes,
		isArchived:        false,
		asOfDate:          asOfDate,
		createdAt:         time.Now(),
	}, nil
}

func ReconstituteLiability(
	id LiabilityID,
	userID UserID,
	name string,
	liabilityType LiabilityType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	isArchived bool,
	asOfDate *time.Time,
	createdAt time.Time,
	debt LiabilityDebtDetails,
) *Liability {
	return &Liability{
		id:                id,
		userID:            userID,
		name:              name,
		liabilityType:     liabilityType,
		currencyID:        currencyID,
		currentBalance:    currentBalance,
		originalPrincipal: debt.OriginalPrincipal,
		interestRateAPR:   debt.InterestRateAPR,
		minimumPayment:    debt.MinimumPayment,
		nextDueDate:       debt.NextDueDate,
		notes:             notes,
		isArchived:        isArchived,
		asOfDate:          asOfDate,
		createdAt:         createdAt,
	}
}

func (l *Liability) AssignID(id LiabilityID)           { l.id = id }
func (l *Liability) ID() LiabilityID                   { return l.id }
func (l *Liability) UserID() UserID                    { return l.userID }
func (l *Liability) Name() string                      { return l.name }
func (l *Liability) Type() LiabilityType               { return l.liabilityType }
func (l *Liability) CurrencyID() CurrencyID            { return l.currencyID }
func (l *Liability) CurrentBalance() float64           { return l.currentBalance }
func (l *Liability) OriginalPrincipal() *float64       { return l.originalPrincipal }
func (l *Liability) InterestRateAPR() *float64         { return l.interestRateAPR }
func (l *Liability) MinimumPayment() float64           { return l.minimumPayment }
func (l *Liability) NextDueDate() *time.Time           { return l.nextDueDate }
func (l *Liability) Notes() string                     { return l.notes }
func (l *Liability) IsArchived() bool                  { return l.isArchived }
func (l *Liability) AsOfDate() *time.Time              { return l.asOfDate }
func (l *Liability) CreatedAt() time.Time              { return l.createdAt }

func (l *Liability) Update(
	name string,
	liabilityType LiabilityType,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
	debt LiabilityDebtDetails,
) error {
	if name == "" {
		return errors.New("liability name cannot be empty")
	}
	if _, err := ParseLiabilityType(string(liabilityType)); err != nil {
		return err
	}
	if currentBalance < 0 {
		return errors.New("current balance cannot be negative")
	}
	if debt.MinimumPayment < 0 {
		return errors.New("minimum payment cannot be negative")
	}
	if debt.OriginalPrincipal != nil && *debt.OriginalPrincipal < 0 {
		return errors.New("original principal cannot be negative")
	}
	if debt.InterestRateAPR != nil && *debt.InterestRateAPR < 0 {
		return errors.New("interest rate cannot be negative")
	}
	l.name = name
	l.liabilityType = liabilityType
	l.currentBalance = currentBalance
	l.originalPrincipal = debt.OriginalPrincipal
	l.interestRateAPR = debt.InterestRateAPR
	l.minimumPayment = debt.MinimumPayment
	l.nextDueDate = debt.NextDueDate
	l.notes = notes
	l.asOfDate = asOfDate
	return nil
}

func (l *Liability) ApplyPayment(amount float64) error {
	if amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}
	if amount > l.currentBalance {
		return errors.New("payment amount exceeds current balance")
	}
	l.currentBalance -= amount
	return nil
}

func (l *Liability) ReversePayment(amount float64) error {
	if amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}
	l.currentBalance += amount
	return nil
}

func (l *Liability) PayoffProgressPercent() *float64 {
	if l.originalPrincipal == nil || *l.originalPrincipal <= 0 {
		return nil
	}
	paid := *l.originalPrincipal - l.currentBalance
	if paid < 0 {
		paid = 0
	}
	pct := (paid / *l.originalPrincipal) * 100
	if pct > 100 {
		pct = 100
	}
	return &pct
}

func (l *Liability) EstimatedMonthsRemaining() *int {
	if l.minimumPayment <= 0 || l.currentBalance <= 0 {
		return nil
	}

	apr := 0.0
	if l.interestRateAPR != nil && *l.interestRateAPR > 0 {
		apr = *l.interestRateAPR
	}

	// Zero APR: classic ceil(balance / payment).
	if apr == 0 {
		months := int((l.currentBalance + l.minimumPayment - 1e-9) / l.minimumPayment)
		if months < 1 {
			months = 1
		}
		return &months
	}

	monthlyRate := apr / 100 / 12
	interestFirst := l.currentBalance * monthlyRate
	if l.minimumPayment <= interestFirst {
		return nil
	}

	// n = log(P / (P - r*B)) / log(1+r) for standard amortizing loan.
	ratio := l.minimumPayment / (l.minimumPayment - interestFirst)
	monthsFloat := math.Log(ratio) / math.Log(1+monthlyRate)
	months := int(math.Ceil(monthsFloat - 1e-9))
	if months < 1 {
		months = 1
	}
	return &months
}

func (l *Liability) Archive()   { l.isArchived = true }
func (l *Liability) Unarchive() { l.isArchived = false }

type LiabilityPaymentID struct{ value int }

func NewLiabilityPaymentID(id int) LiabilityPaymentID { return LiabilityPaymentID{value: id} }
func (p LiabilityPaymentID) Value() int               { return p.value }

type LiabilityPayment struct {
	id          LiabilityPaymentID
	liabilityID LiabilityID
	userID      UserID
	amount      float64
	paidAt      time.Time
	expenseID   *int
	note        string
	createdAt   time.Time
}

func NewLiabilityPayment(
	liabilityID LiabilityID,
	userID UserID,
	amount float64,
	paidAt time.Time,
	expenseID *int,
	note string,
) (*LiabilityPayment, error) {
	if amount <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}
	return &LiabilityPayment{
		liabilityID: liabilityID,
		userID:      userID,
		amount:      amount,
		paidAt:      paidAt,
		expenseID:   expenseID,
		note:        note,
		createdAt:   time.Now(),
	}, nil
}

func ReconstituteLiabilityPayment(
	id LiabilityPaymentID,
	liabilityID LiabilityID,
	userID UserID,
	amount float64,
	paidAt time.Time,
	expenseID *int,
	note string,
	createdAt time.Time,
) *LiabilityPayment {
	return &LiabilityPayment{
		id:          id,
		liabilityID: liabilityID,
		userID:      userID,
		amount:      amount,
		paidAt:      paidAt,
		expenseID:   expenseID,
		note:        note,
		createdAt:   createdAt,
	}
}

func (p *LiabilityPayment) AssignID(id LiabilityPaymentID) { p.id = id }
func (p *LiabilityPayment) ID() LiabilityPaymentID         { return p.id }
func (p *LiabilityPayment) LiabilityID() LiabilityID       { return p.liabilityID }
func (p *LiabilityPayment) UserID() UserID                 { return p.userID }
func (p *LiabilityPayment) Amount() float64                { return p.amount }
func (p *LiabilityPayment) PaidAt() time.Time              { return p.paidAt }
func (p *LiabilityPayment) ExpenseID() *int                { return p.expenseID }
func (p *LiabilityPayment) Note() string                   { return p.note }
func (p *LiabilityPayment) CreatedAt() time.Time           { return p.createdAt }

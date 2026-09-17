package finance

import (
	"errors"
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
	id             LiabilityID
	userID         UserID
	name           string
	liabilityType  LiabilityType
	currencyID     CurrencyID
	currentBalance float64
	notes          string
	isArchived     bool
	asOfDate       *time.Time
	createdAt      time.Time
}

func NewLiability(
	userID UserID,
	name string,
	liabilityType LiabilityType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
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
	return &Liability{
		userID:         userID,
		name:           name,
		liabilityType:  liabilityType,
		currencyID:     currencyID,
		currentBalance: currentBalance,
		notes:          notes,
		isArchived:     false,
		asOfDate:       asOfDate,
		createdAt:      time.Now(),
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
) *Liability {
	return &Liability{
		id:             id,
		userID:         userID,
		name:           name,
		liabilityType:  liabilityType,
		currencyID:     currencyID,
		currentBalance: currentBalance,
		notes:          notes,
		isArchived:     isArchived,
		asOfDate:       asOfDate,
		createdAt:      createdAt,
	}
}

func (l *Liability) AssignID(id LiabilityID)      { l.id = id }
func (l *Liability) ID() LiabilityID              { return l.id }
func (l *Liability) UserID() UserID               { return l.userID }
func (l *Liability) Name() string                 { return l.name }
func (l *Liability) Type() LiabilityType          { return l.liabilityType }
func (l *Liability) CurrencyID() CurrencyID       { return l.currencyID }
func (l *Liability) CurrentBalance() float64      { return l.currentBalance }
func (l *Liability) Notes() string                { return l.notes }
func (l *Liability) IsArchived() bool             { return l.isArchived }
func (l *Liability) AsOfDate() *time.Time         { return l.asOfDate }
func (l *Liability) CreatedAt() time.Time         { return l.createdAt }

func (l *Liability) Update(name string, liabilityType LiabilityType, currentBalance float64, notes string, asOfDate *time.Time) error {
	if name == "" {
		return errors.New("liability name cannot be empty")
	}
	if _, err := ParseLiabilityType(string(liabilityType)); err != nil {
		return err
	}
	if currentBalance < 0 {
		return errors.New("current balance cannot be negative")
	}
	l.name = name
	l.liabilityType = liabilityType
	l.currentBalance = currentBalance
	l.notes = notes
	l.asOfDate = asOfDate
	return nil
}

func (l *Liability) Archive()   { l.isArchived = true }
func (l *Liability) Unarchive() { l.isArchived = false }

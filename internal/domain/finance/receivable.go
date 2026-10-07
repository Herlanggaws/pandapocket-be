package finance

import (
	"errors"
	"time"
)

type ReceivableType string

const (
	ReceivableTypePersonalLoan ReceivableType = "personal_loan"
	ReceivableTypeInvoice      ReceivableType = "invoice"
	ReceivableTypeOther        ReceivableType = "other"
)

func ParseReceivableType(value string) (ReceivableType, error) {
	switch ReceivableType(value) {
	case ReceivableTypePersonalLoan, ReceivableTypeInvoice, ReceivableTypeOther:
		return ReceivableType(value), nil
	default:
		return "", errors.New("invalid receivable type")
	}
}

type ReceivableID struct{ value int }

func NewReceivableID(id int) ReceivableID { return ReceivableID{value: id} }
func (r ReceivableID) Value() int         { return r.value }

type Receivable struct {
	id                 ReceivableID
	userID             UserID
	name               string
	receivableType     ReceivableType
	currencyID         CurrencyID
	currentBalance     float64
	originalPrincipal  *float64
	nextDueDate        *time.Time
	notes              string
	isArchived         bool
	asOfDate           *time.Time
	createExpenseID    *int
	createdAt          time.Time
}

type ReceivableDetails struct {
	OriginalPrincipal *float64
	NextDueDate       *time.Time
	CreateExpenseID   *int
}

func NewReceivable(
	userID UserID,
	name string,
	receivableType ReceivableType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
	details ReceivableDetails,
) (*Receivable, error) {
	if name == "" {
		return nil, errors.New("receivable name cannot be empty")
	}
	if _, err := ParseReceivableType(string(receivableType)); err != nil {
		return nil, err
	}
	if currentBalance < 0 {
		return nil, errors.New("current balance cannot be negative")
	}
	if details.OriginalPrincipal != nil && *details.OriginalPrincipal < 0 {
		return nil, errors.New("original principal cannot be negative")
	}
	return &Receivable{
		userID:            userID,
		name:              name,
		receivableType:    receivableType,
		currencyID:        currencyID,
		currentBalance:    currentBalance,
		originalPrincipal: details.OriginalPrincipal,
		nextDueDate:       details.NextDueDate,
		notes:             notes,
		isArchived:        false,
		asOfDate:          asOfDate,
		createExpenseID:   details.CreateExpenseID,
		createdAt:         time.Now(),
	}, nil
}

func ReconstituteReceivable(
	id ReceivableID,
	userID UserID,
	name string,
	receivableType ReceivableType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	isArchived bool,
	asOfDate *time.Time,
	createdAt time.Time,
	details ReceivableDetails,
) *Receivable {
	return &Receivable{
		id:                id,
		userID:            userID,
		name:              name,
		receivableType:    receivableType,
		currencyID:        currencyID,
		currentBalance:    currentBalance,
		originalPrincipal: details.OriginalPrincipal,
		nextDueDate:       details.NextDueDate,
		notes:             notes,
		isArchived:        isArchived,
		asOfDate:          asOfDate,
		createExpenseID:   details.CreateExpenseID,
		createdAt:         createdAt,
	}
}

func (r *Receivable) AssignID(id ReceivableID)           { r.id = id }
func (r *Receivable) ID() ReceivableID                   { return r.id }
func (r *Receivable) UserID() UserID                     { return r.userID }
func (r *Receivable) Name() string                       { return r.name }
func (r *Receivable) Type() ReceivableType               { return r.receivableType }
func (r *Receivable) CurrencyID() CurrencyID             { return r.currencyID }
func (r *Receivable) CurrentBalance() float64            { return r.currentBalance }
func (r *Receivable) OriginalPrincipal() *float64        { return r.originalPrincipal }
func (r *Receivable) NextDueDate() *time.Time            { return r.nextDueDate }
func (r *Receivable) Notes() string                      { return r.notes }
func (r *Receivable) IsArchived() bool                   { return r.isArchived }
func (r *Receivable) AsOfDate() *time.Time               { return r.asOfDate }
func (r *Receivable) CreateExpenseID() *int              { return r.createExpenseID }
func (r *Receivable) CreatedAt() time.Time               { return r.createdAt }

func (r *Receivable) SetCreateExpenseID(expenseID *int) { r.createExpenseID = expenseID }

func (r *Receivable) Update(
	name string,
	receivableType ReceivableType,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
	details ReceivableDetails,
) error {
	if name == "" {
		return errors.New("receivable name cannot be empty")
	}
	if _, err := ParseReceivableType(string(receivableType)); err != nil {
		return err
	}
	if currentBalance < 0 {
		return errors.New("current balance cannot be negative")
	}
	if details.OriginalPrincipal != nil && *details.OriginalPrincipal < 0 {
		return errors.New("original principal cannot be negative")
	}
	r.name = name
	r.receivableType = receivableType
	r.currentBalance = currentBalance
	r.originalPrincipal = details.OriginalPrincipal
	r.nextDueDate = details.NextDueDate
	r.notes = notes
	r.asOfDate = asOfDate
	return nil
}

func (r *Receivable) ApplyCollection(amount float64) error {
	if amount <= 0 {
		return errors.New("collection amount must be greater than zero")
	}
	if amount > r.currentBalance {
		return errors.New("collection amount exceeds current balance")
	}
	r.currentBalance -= amount
	return nil
}

func (r *Receivable) ReverseCollection(amount float64) error {
	if amount <= 0 {
		return errors.New("collection amount must be greater than zero")
	}
	r.currentBalance += amount
	return nil
}

func (r *Receivable) CollectionProgressPercent() *float64 {
	if r.originalPrincipal == nil || *r.originalPrincipal <= 0 {
		return nil
	}
	collected := *r.originalPrincipal - r.currentBalance
	if collected < 0 {
		collected = 0
	}
	pct := (collected / *r.originalPrincipal) * 100
	if pct > 100 {
		pct = 100
	}
	return &pct
}

func (r *Receivable) Archive()   { r.isArchived = true }
func (r *Receivable) Unarchive() { r.isArchived = false }

type ReceivableCollectionID struct{ value int }

func NewReceivableCollectionID(id int) ReceivableCollectionID {
	return ReceivableCollectionID{value: id}
}
func (c ReceivableCollectionID) Value() int { return c.value }

type ReceivableCollection struct {
	id           ReceivableCollectionID
	receivableID ReceivableID
	userID       UserID
	amount       float64
	collectedAt  time.Time
	incomeID     *int
	note         string
	createdAt    time.Time
}

func NewReceivableCollection(
	receivableID ReceivableID,
	userID UserID,
	amount float64,
	collectedAt time.Time,
	incomeID *int,
	note string,
) (*ReceivableCollection, error) {
	if amount <= 0 {
		return nil, errors.New("collection amount must be greater than zero")
	}
	return &ReceivableCollection{
		receivableID: receivableID,
		userID:       userID,
		amount:       amount,
		collectedAt:  collectedAt,
		incomeID:     incomeID,
		note:         note,
		createdAt:    time.Now(),
	}, nil
}

func ReconstituteReceivableCollection(
	id ReceivableCollectionID,
	receivableID ReceivableID,
	userID UserID,
	amount float64,
	collectedAt time.Time,
	incomeID *int,
	note string,
	createdAt time.Time,
) *ReceivableCollection {
	return &ReceivableCollection{
		id:           id,
		receivableID: receivableID,
		userID:       userID,
		amount:       amount,
		collectedAt:  collectedAt,
		incomeID:     incomeID,
		note:         note,
		createdAt:    createdAt,
	}
}

func (c *ReceivableCollection) AssignID(id ReceivableCollectionID) { c.id = id }
func (c *ReceivableCollection) ID() ReceivableCollectionID         { return c.id }
func (c *ReceivableCollection) ReceivableID() ReceivableID         { return c.receivableID }
func (c *ReceivableCollection) UserID() UserID                     { return c.userID }
func (c *ReceivableCollection) Amount() float64                    { return c.amount }
func (c *ReceivableCollection) CollectedAt() time.Time             { return c.collectedAt }
func (c *ReceivableCollection) IncomeID() *int                     { return c.incomeID }
func (c *ReceivableCollection) Note() string                       { return c.note }
func (c *ReceivableCollection) CreatedAt() time.Time               { return c.createdAt }

package ticket

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxSubjectLength = 200
	MaxBodyLength    = 2000
)

var (
	ErrNotFound     = errors.New("ticket not found")
	ErrAccessDenied = errors.New("access denied to ticket")
	ErrCannotReopen = errors.New("only done tickets can be reopened")
)

type Category string

const (
	CategoryTechnical Category = "technical"
	CategoryPayment   Category = "payment"
	CategoryOther     Category = "other"
)

func ParseCategory(value string) (Category, error) {
	switch Category(strings.TrimSpace(value)) {
	case CategoryTechnical:
		return CategoryTechnical, nil
	case CategoryPayment:
		return CategoryPayment, nil
	case CategoryOther:
		return CategoryOther, nil
	default:
		return "", errors.New("category must be technical, payment, or other")
	}
}

func (c Category) String() string { return string(c) }

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

func ParsePriority(value string) (Priority, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return PriorityMedium, nil
	}
	switch Priority(trimmed) {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return Priority(trimmed), nil
	default:
		return "", errors.New("priority must be low, medium, or high")
	}
}

func (p Priority) String() string { return string(p) }

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func ParseStatus(value string) (Status, error) {
	switch Status(strings.TrimSpace(value)) {
	case StatusOpen, StatusInProgress, StatusDone:
		return Status(strings.TrimSpace(value)), nil
	default:
		return "", errors.New("status must be open, in_progress, or done")
	}
}

func (s Status) String() string { return string(s) }

type TicketID struct {
	value int
}

func NewTicketID(id int) TicketID {
	return TicketID{value: id}
}

func (id TicketID) Value() int { return id.value }

type Ticket struct {
	id        TicketID
	userID    int
	subject   string
	body      string
	category  Category
	priority  Priority
	status    Status
	createdAt time.Time
	updatedAt time.Time
}

func NewTicket(userID int, subject string, body string, category Category, priority Priority) (*Ticket, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}

	trimmedSubject := strings.TrimSpace(subject)
	if trimmedSubject == "" {
		return nil, errors.New("subject is required")
	}
	if utf8.RuneCountInString(trimmedSubject) > MaxSubjectLength {
		return nil, errors.New("subject must be at most 200 characters")
	}

	trimmedBody := strings.TrimSpace(body)
	if trimmedBody == "" {
		return nil, errors.New("body is required")
	}
	if utf8.RuneCountInString(trimmedBody) > MaxBodyLength {
		return nil, errors.New("body must be at most 2000 characters")
	}

	now := time.Now()
	return &Ticket{
		userID:    userID,
		subject:   trimmedSubject,
		body:      trimmedBody,
		category:  category,
		priority:  priority,
		status:    StatusOpen,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func ReconstituteTicket(
	id TicketID,
	userID int,
	subject, body string,
	category Category,
	priority Priority,
	status Status,
	createdAt, updatedAt time.Time,
) *Ticket {
	return &Ticket{
		id:        id,
		userID:    userID,
		subject:   subject,
		body:      body,
		category:  category,
		priority:  priority,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t *Ticket) ID() TicketID         { return t.id }
func (t *Ticket) UserID() int          { return t.userID }
func (t *Ticket) Subject() string      { return t.subject }
func (t *Ticket) Body() string         { return t.body }
func (t *Ticket) Category() Category   { return t.category }
func (t *Ticket) Priority() Priority   { return t.priority }
func (t *Ticket) Status() Status       { return t.status }
func (t *Ticket) CreatedAt() time.Time { return t.createdAt }
func (t *Ticket) UpdatedAt() time.Time { return t.updatedAt }

func (t *Ticket) AssignID(id TicketID) {
	t.id = id
}

func (t *Ticket) UpdateStatus(status Status) error {
	if status != StatusOpen && status != StatusInProgress && status != StatusDone {
		return errors.New("status must be open, in_progress, or done")
	}
	t.status = status
	t.updatedAt = time.Now()
	return nil
}

func (t *Ticket) Reopen() error {
	if t.status != StatusDone {
		return ErrCannotReopen
	}
	t.status = StatusOpen
	t.updatedAt = time.Now()
	return nil
}

type ListFilter struct {
	UserID   *int
	Status   *Status
	Category *Category
}

type TicketWithUser struct {
	Ticket *Ticket
	Email  string
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *Ticket) error
	FindByID(ctx context.Context, id TicketID) (*Ticket, error)
	List(ctx context.Context, filter ListFilter) ([]*Ticket, error)
	ListWithUser(ctx context.Context, filter ListFilter) ([]TicketWithUser, error)
	Update(ctx context.Context, ticket *Ticket) error
}

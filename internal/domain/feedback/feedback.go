package feedback

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxMessageLength = 2000

type Category string

const (
	CategoryBug        Category = "bug"
	CategorySuggestion Category = "suggestion"
	CategoryOther      Category = "other"
)

func ParseCategory(value string) (Category, error) {
	switch Category(strings.TrimSpace(value)) {
	case CategoryBug:
		return CategoryBug, nil
	case CategorySuggestion:
		return CategorySuggestion, nil
	case CategoryOther:
		return CategoryOther, nil
	default:
		return "", errors.New("category must be bug, suggestion, or other")
	}
}

func (c Category) String() string {
	return string(c)
}

type FeedbackID struct {
	value int
}

func NewFeedbackID(id int) FeedbackID {
	return FeedbackID{value: id}
}

func (id FeedbackID) Value() int {
	return id.value
}

type Feedback struct {
	id        FeedbackID
	userID    int
	category  Category
	message   string
	createdAt time.Time
	updatedAt time.Time
}

func NewFeedback(userID int, category Category, message string) (*Feedback, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return nil, errors.New("message is required")
	}
	if utf8.RuneCountInString(trimmed) > MaxMessageLength {
		return nil, errors.New("message must be at most 2000 characters")
	}
	now := time.Now()
	return &Feedback{
		userID:    userID,
		category:  category,
		message:   trimmed,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func ReconstituteFeedback(
	id FeedbackID,
	userID int,
	category Category,
	message string,
	createdAt, updatedAt time.Time,
) *Feedback {
	return &Feedback{
		id:        id,
		userID:    userID,
		category:  category,
		message:   message,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (f *Feedback) ID() FeedbackID      { return f.id }
func (f *Feedback) UserID() int         { return f.userID }
func (f *Feedback) Category() Category  { return f.category }
func (f *Feedback) Message() string     { return f.message }
func (f *Feedback) CreatedAt() time.Time { return f.createdAt }
func (f *Feedback) UpdatedAt() time.Time { return f.updatedAt }

func (f *Feedback) AssignID(id FeedbackID) {
	f.id = id
}

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *Feedback) error
}

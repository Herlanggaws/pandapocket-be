package database

import (
	"context"
	"errors"

	"panda-pocket/internal/domain/ticket"

	"gorm.io/gorm"
)

type GormTicketRepository struct {
	db *gorm.DB
}

func NewGormTicketRepository(db *gorm.DB) *GormTicketRepository {
	return &GormTicketRepository{db: db}
}

func (r *GormTicketRepository) Create(ctx context.Context, item *ticket.Ticket) error {
	model := &SupportTicket{
		UserID:   uint(item.UserID()),
		Subject:  item.Subject(),
		Body:     item.Body(),
		Category: item.Category().String(),
		Priority: item.Priority().String(),
		Status:   item.Status().String(),
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	item.AssignID(ticket.NewTicketID(int(model.ID)))
	return nil
}

func (r *GormTicketRepository) FindByID(ctx context.Context, id ticket.TicketID) (*ticket.Ticket, error) {
	var model SupportTicket
	if err := r.db.WithContext(ctx).First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ticket.ErrNotFound
		}
		return nil, err
	}
	return mapSupportTicketToDomain(&model)
}

func (r *GormTicketRepository) List(ctx context.Context, filter ticket.ListFilter) ([]*ticket.Ticket, error) {
	query := r.db.WithContext(ctx).Model(&SupportTicket{})
	query = applyTicketFilter(query, filter)

	var models []SupportTicket
	if err := query.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]*ticket.Ticket, 0, len(models))
	for i := range models {
		mapped, err := mapSupportTicketToDomain(&models[i])
		if err != nil {
			return nil, err
		}
		items = append(items, mapped)
	}
	return items, nil
}

func (r *GormTicketRepository) ListWithUser(ctx context.Context, filter ticket.ListFilter) ([]ticket.TicketWithUser, error) {
	query := r.db.WithContext(ctx).Model(&SupportTicket{}).Preload("User")
	query = applyTicketFilter(query, filter)

	var models []SupportTicket
	if err := query.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]ticket.TicketWithUser, 0, len(models))
	for i := range models {
		mapped, err := mapSupportTicketToDomain(&models[i])
		if err != nil {
			return nil, err
		}
		email := ""
		if models[i].User != nil {
			email = models[i].User.Email
		}
		items = append(items, ticket.TicketWithUser{Ticket: mapped, Email: email})
	}
	return items, nil
}

func (r *GormTicketRepository) Update(ctx context.Context, item *ticket.Ticket) error {
	result := r.db.WithContext(ctx).Model(&SupportTicket{}).
		Where("id = ?", item.ID().Value()).
		Updates(map[string]interface{}{
			"status":     item.Status().String(),
			"updated_at": item.UpdatedAt(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ticket.ErrNotFound
	}
	return nil
}

func applyTicketFilter(query *gorm.DB, filter ticket.ListFilter) *gorm.DB {
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", filter.Status.String())
	}
	if filter.Category != nil {
		query = query.Where("category = ?", filter.Category.String())
	}
	return query
}

func mapSupportTicketToDomain(model *SupportTicket) (*ticket.Ticket, error) {
	category, err := ticket.ParseCategory(model.Category)
	if err != nil {
		return nil, err
	}
	priority, err := ticket.ParsePriority(model.Priority)
	if err != nil {
		return nil, err
	}
	status, err := ticket.ParseStatus(model.Status)
	if err != nil {
		return nil, err
	}
	return ticket.ReconstituteTicket(
		ticket.NewTicketID(int(model.ID)),
		int(model.UserID),
		model.Subject,
		model.Body,
		category,
		priority,
		status,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

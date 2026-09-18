package ticket

import (
	"context"
	"time"

	"panda-pocket/internal/domain/entitlement"
	domainTicket "panda-pocket/internal/domain/ticket"
)

type CreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required,min=1,max=200"`
	Body     string `json:"body" binding:"required,min=1,max=2000"`
	Category string `json:"category" binding:"required,oneof=technical payment other"`
	Priority string `json:"priority" binding:"omitempty,oneof=low medium high"`
}

type TicketResponse struct {
	ID        int    `json:"id"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Category  string `json:"category"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AdminTicketResponse struct {
	TicketResponse
	UserID    int    `json:"user_id"`
	UserEmail string `json:"user_email,omitempty"`
}

func ToTicketResponse(item *domainTicket.Ticket) TicketResponse {
	return TicketResponse{
		ID:        item.ID().Value(),
		Subject:   item.Subject(),
		Body:      item.Body(),
		Category:  item.Category().String(),
		Priority:  item.Priority().String(),
		Status:    item.Status().String(),
		CreatedAt: item.CreatedAt().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt().Format(time.RFC3339),
	}
}

type CreateTicketUseCase struct {
	repo         domainTicket.TicketRepository
	entitlements entitlement.Checker
}

func NewCreateTicketUseCase(
	repo domainTicket.TicketRepository,
	entitlements entitlement.Checker,
) *CreateTicketUseCase {
	return &CreateTicketUseCase{repo: repo, entitlements: entitlements}
}

func (uc *CreateTicketUseCase) Execute(ctx context.Context, userID int, req CreateTicketRequest) (*TicketResponse, error) {
	isPro, err := uc.entitlements.IsPro(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !isPro {
		return nil, entitlement.RequirePro(entitlement.FeatureTickets)
	}

	category, err := domainTicket.ParseCategory(req.Category)
	if err != nil {
		return nil, err
	}
	priority, err := domainTicket.ParsePriority(req.Priority)
	if err != nil {
		return nil, err
	}

	item, err := domainTicket.NewTicket(userID, req.Subject, req.Body, category, priority)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	response := ToTicketResponse(item)
	return &response, nil
}

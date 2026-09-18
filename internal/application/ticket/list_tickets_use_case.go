package ticket

import (
	"context"

	domainTicket "panda-pocket/internal/domain/ticket"
)

type ListTicketsUseCase struct {
	repo domainTicket.TicketRepository
}

func NewListTicketsUseCase(repo domainTicket.TicketRepository) *ListTicketsUseCase {
	return &ListTicketsUseCase{repo: repo}
}

func (uc *ListTicketsUseCase) Execute(ctx context.Context, userID int, statusFilter, categoryFilter string) ([]TicketResponse, error) {
	filter := domainTicket.ListFilter{UserID: &userID}
	if statusFilter != "" {
		status, err := domainTicket.ParseStatus(statusFilter)
		if err != nil {
			return nil, err
		}
		filter.Status = &status
	}
	if categoryFilter != "" {
		category, err := domainTicket.ParseCategory(categoryFilter)
		if err != nil {
			return nil, err
		}
		filter.Category = &category
	}

	items, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]TicketResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, ToTicketResponse(item))
	}
	return responses, nil
}

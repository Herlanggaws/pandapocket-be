package ticket

import (
	"context"
	"strconv"

	domainTicket "panda-pocket/internal/domain/ticket"
)

type ListAdminTicketsUseCase struct {
	repo domainTicket.TicketRepository
}

func NewListAdminTicketsUseCase(repo domainTicket.TicketRepository) *ListAdminTicketsUseCase {
	return &ListAdminTicketsUseCase{repo: repo}
}

func (uc *ListAdminTicketsUseCase) Execute(
	ctx context.Context,
	statusFilter, categoryFilter, userIDFilter string,
) ([]AdminTicketResponse, error) {
	filter := domainTicket.ListFilter{}
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
	if userIDFilter != "" {
		userID, err := strconv.Atoi(userIDFilter)
		if err != nil || userID <= 0 {
			return nil, domainTicket.ErrNotFound
		}
		filter.UserID = &userID
	}

	items, err := uc.repo.ListWithUser(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]AdminTicketResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, AdminTicketResponse{
			TicketResponse: ToTicketResponse(item.Ticket),
			UserID:         item.Ticket.UserID(),
			UserEmail:      item.Email,
		})
	}
	return responses, nil
}

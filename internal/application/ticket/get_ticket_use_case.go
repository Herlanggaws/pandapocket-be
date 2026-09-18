package ticket

import (
	"context"
	"strconv"

	domainTicket "panda-pocket/internal/domain/ticket"
)

type GetTicketUseCase struct {
	repo domainTicket.TicketRepository
}

func NewGetTicketUseCase(repo domainTicket.TicketRepository) *GetTicketUseCase {
	return &GetTicketUseCase{repo: repo}
}

func (uc *GetTicketUseCase) Execute(ctx context.Context, userID int, ticketIDStr string) (*TicketResponse, error) {
	ticketID, err := parseTicketID(ticketIDStr)
	if err != nil {
		return nil, err
	}

	item, err := uc.repo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if item.UserID() != userID {
		return nil, domainTicket.ErrAccessDenied
	}

	response := ToTicketResponse(item)
	return &response, nil
}

func parseTicketID(value string) (domainTicket.TicketID, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return domainTicket.TicketID{}, domainTicket.ErrNotFound
	}
	return domainTicket.NewTicketID(id), nil
}

package ticket

import (
	"context"

	"panda-pocket/internal/domain/identity"
	domainTicket "panda-pocket/internal/domain/ticket"
)

type GetAdminTicketUseCase struct {
	repo     domainTicket.TicketRepository
	userRepo identity.UserRepository
}

func NewGetAdminTicketUseCase(
	repo domainTicket.TicketRepository,
	userRepo identity.UserRepository,
) *GetAdminTicketUseCase {
	return &GetAdminTicketUseCase{repo: repo, userRepo: userRepo}
}

func (uc *GetAdminTicketUseCase) Execute(ctx context.Context, ticketIDStr string) (*AdminTicketResponse, error) {
	ticketID, err := parseTicketID(ticketIDStr)
	if err != nil {
		return nil, err
	}

	item, err := uc.repo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	email := ""
	if user, err := uc.userRepo.FindByID(ctx, identity.NewUserID(item.UserID())); err == nil {
		email = user.Email().Value()
	}

	response := AdminTicketResponse{
		TicketResponse: ToTicketResponse(item),
		UserID:         item.UserID(),
		UserEmail:      email,
	}
	return &response, nil
}

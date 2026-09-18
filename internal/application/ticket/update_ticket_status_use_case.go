package ticket

import (
	"context"
	"fmt"

	"panda-pocket/internal/domain/identity"
	"panda-pocket/internal/domain/notification"
	domainTicket "panda-pocket/internal/domain/ticket"
)

type UpdateTicketStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open in_progress done"`
}

type UpdateTicketStatusUseCase struct {
	repo         domainTicket.TicketRepository
	userRepo     identity.UserRepository
	emailService notification.EmailService
}

func NewUpdateTicketStatusUseCase(
	repo domainTicket.TicketRepository,
	userRepo identity.UserRepository,
	emailService notification.EmailService,
) *UpdateTicketStatusUseCase {
	return &UpdateTicketStatusUseCase{
		repo:         repo,
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (uc *UpdateTicketStatusUseCase) Execute(
	ctx context.Context,
	ticketIDStr string,
	req UpdateTicketStatusRequest,
) (*AdminTicketResponse, error) {
	ticketID, err := parseTicketID(ticketIDStr)
	if err != nil {
		return nil, err
	}

	status, err := domainTicket.ParseStatus(req.Status)
	if err != nil {
		return nil, err
	}

	item, err := uc.repo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if err := item.UpdateStatus(status); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	email := ""
	user, err := uc.userRepo.FindByID(ctx, identity.NewUserID(item.UserID()))
	if err != nil {
		fmt.Println("Failed to load user for ticket email:", err)
	} else {
		email = user.Email().Value()
		go sendTicketStatusEmail(uc.emailService, email, item)
	}

	response := AdminTicketResponse{
		TicketResponse: ToTicketResponse(item),
		UserID:         item.UserID(),
		UserEmail:      email,
	}
	return &response, nil
}

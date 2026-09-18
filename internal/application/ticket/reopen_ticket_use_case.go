package ticket

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"

	"panda-pocket/internal/domain/identity"
	"panda-pocket/internal/domain/notification"
	domainTicket "panda-pocket/internal/domain/ticket"
)

type ReopenTicketUseCase struct {
	repo         domainTicket.TicketRepository
	userRepo     identity.UserRepository
	emailService notification.EmailService
}

func NewReopenTicketUseCase(
	repo domainTicket.TicketRepository,
	userRepo identity.UserRepository,
	emailService notification.EmailService,
) *ReopenTicketUseCase {
	return &ReopenTicketUseCase{
		repo:         repo,
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (uc *ReopenTicketUseCase) Execute(ctx context.Context, userID int, ticketIDStr string) (*TicketResponse, error) {
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

	if err := item.Reopen(); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	uc.sendStatusEmail(ctx, item)

	response := ToTicketResponse(item)
	return &response, nil
}

func (uc *ReopenTicketUseCase) sendStatusEmail(ctx context.Context, item *domainTicket.Ticket) {
	user, err := uc.userRepo.FindByID(ctx, identity.NewUserID(item.UserID()))
	if err != nil {
		fmt.Println("Failed to load user for ticket email:", err)
		return
	}

	go sendTicketStatusEmail(uc.emailService, user.Email().Value(), item)
}

func sendTicketStatusEmail(emailService notification.EmailService, to string, item *domainTicket.Ticket) {
	tmpl, err := template.ParseFiles("internal/infrastructure/notification/templates/ticket_status_changed.html")
	if err != nil {
		fmt.Println("Failed to parse ticket email template:", err)
		return
	}

	appURL := os.Getenv("APP_URL")
	ticketURL := fmt.Sprintf("%s/help/tickets/%d", appURL, item.ID().Value())

	var body bytes.Buffer
	data := struct {
		TicketID  int
		Subject   string
		Status    string
		TicketURL string
	}{
		TicketID:  item.ID().Value(),
		Subject:   item.Subject(),
		Status:    item.Status().String(),
		TicketURL: ticketURL,
	}
	if err := tmpl.Execute(&body, data); err != nil {
		fmt.Println("Failed to render ticket email template:", err)
		return
	}

	emailMsg := notification.EmailMessage{
		To:      to,
		Subject: fmt.Sprintf("Ticket #%d status: %s", item.ID().Value(), item.Status().String()),
		Body:    body.String(),
	}
	if err := emailService.SendEmail(context.Background(), emailMsg); err != nil {
		fmt.Println("Failed to send ticket status email:", err)
	}
}

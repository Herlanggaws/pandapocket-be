package identity

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"os"
	"time"

	"panda-pocket/internal/domain/identity"
	"panda-pocket/internal/domain/notification"
)

// ForgotPasswordRequest represents the request to check for forgot password email
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPasswordResponse represents the response for forgot password
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ForgotPasswordUseCase handles forgot password logic
type ForgotPasswordUseCase struct {
	userRepository  identity.UserRepository
	tokenRepository identity.PasswordResetTokenRepository
	emailService    notification.EmailService
}

// NewForgotPasswordUseCase creates a new forgot password use case
func NewForgotPasswordUseCase(
	userRepository identity.UserRepository,
	tokenRepository identity.PasswordResetTokenRepository,
	emailService notification.EmailService,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		emailService:    emailService,
	}
}

const forgotPasswordGenericMessage = "If an account exists for that email, a password reset link has been sent."

// Execute executes the forgot password use case
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, req ForgotPasswordRequest) (*ForgotPasswordResponse, error) {
	generic := &ForgotPasswordResponse{Message: forgotPasswordGenericMessage}

	email, err := identity.NewEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email format")
	}

	user, err := uc.userRepository.FindByEmail(ctx, email)
	if err != nil {
		return generic, nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.New("failed to generate token")
	}
	tokenString := hex.EncodeToString(tokenBytes)

	token := identity.NewPasswordResetToken(user.ID(), tokenString, time.Now().Add(1*time.Hour))
	if err := uc.tokenRepository.Save(ctx, token); err != nil {
		return nil, errors.New("failed to save token")
	}

	appURL := os.Getenv("APP_URL")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", appURL, tokenString)

	go func() {
		tmpl, err := template.ParseFiles("internal/infrastructure/notification/templates/reset_password.html")
		if err != nil {
			fmt.Println("Failed to parse email template: ", err)
			return
		}

		var body bytes.Buffer
		if err := tmpl.Execute(&body, struct{ ResetLink string }{ResetLink: resetLink}); err != nil {
			fmt.Println("Failed to render email template with err: ", err)
			return
		}

		emailMsg := notification.EmailMessage{
			To:      user.Email().Value(),
			Subject: "Reset Your Password - Berbudget",
			Body:    body.String(),
		}

		if err := uc.emailService.SendEmail(ctx, emailMsg); err != nil {
			fmt.Println("Failed to send reset email with err: ", err)
		}
	}()

	return generic, nil
}

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
	Token   string `json:"token"`
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

// Execute executes the forgot password use case
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, req ForgotPasswordRequest) (*ForgotPasswordResponse, error) {
	// Validate email format
	email, err := identity.NewEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email format")
	}

	// Check if email exists
	user, err := uc.userRepository.FindByEmail(ctx, email)
	if err != nil {
		// If user not found, we still want to return generalized error or silence?
		// User requirement said: "IF email was found, return response 'Email found'"
		// So if NOT found, I should return error.
		return nil, errors.New("email not found")
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.New("failed to generate token")
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Create token entity (expires in 1 hour)
	token := identity.NewPasswordResetToken(user.ID(), tokenString, time.Now().Add(1*time.Hour))

	// Save token
	if err := uc.tokenRepository.Save(ctx, token); err != nil {
		return nil, errors.New("failed to save token")
	}

	// Send email
	appURL := os.Getenv("APP_URL")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", appURL, tokenString)

	// Read and parse template
	// Assuming running from project root
	tmpl, err := template.ParseFiles("internal/infrastructure/notification/templates/reset_password.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, struct{ ResetLink string }{ResetLink: resetLink}); err != nil {
		return nil, fmt.Errorf("failed to render email template: %w", err)
	}

	emailMsg := notification.EmailMessage{
		To:      user.Email().Value(),
		Subject: "Reset Your Password - PandaPocket",
		Body:    body.String(),
	}

	// We don't return error if email fails, just log it (in a real app).
	// Or we can return error depending on requirements.
	// For now let's just log it if we had a logger, but since we don't have a logger injected,
	// we will just try to send it. If it fails, maybe we should let the user know?
	// The user requirement didn't specify error handling for email failure.
	// But usually, if email fails, the process fails.
	if err := uc.emailService.SendEmail(ctx, emailMsg); err != nil {
		// Log error but generally we might not want to block execution if it's async,
		// but here it is synchronous.
		// For now, let's treat it as non-fatal but note it, or just return error.
		// Returning error is safer to let user retry.
		fmt.Println("Failed to send reset email with err: ", err)
		return nil, errors.New("failed to send reset email")
	}

	return &ForgotPasswordResponse{
		Message: "Forgot password successful",
		Token:   tokenString,
	}, nil
}

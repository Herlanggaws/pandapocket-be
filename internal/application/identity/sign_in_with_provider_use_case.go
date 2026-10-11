package identity

import (
	"context"
	"errors"
	"strings"

	"panda-pocket/internal/domain/billing"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type SignInWithProviderRequest struct {
	Provider string `json:"provider" binding:"required"`
	IDToken  string `json:"id_token" binding:"required"`
	Nonce    string `json:"nonce" binding:"required"`
}

type VerifiedSignIn struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
}

type SignInTokenVerifier interface {
	Verify(ctx context.Context, provider, idToken, nonce string) (VerifiedSignIn, error)
}

type SignInWithProviderUseCase struct {
	verifier         SignInTokenVerifier
	users            domainIdentity.UserRepository
	identities       domainIdentity.IdentityRepository
	userService      *domainIdentity.UserService
	tokenService     sessionIssuer
	subscriptionRepo trialStarter
}

type trialStarter interface {
	Save(ctx context.Context, sub *billing.Subscription) error
}

type sessionIssuer interface {
	GenerateToken(ctx context.Context, userID int, email, role string) (string, string, error)
}

func NewSignInWithProviderUseCase(
	verifier SignInTokenVerifier,
	users domainIdentity.UserRepository,
	identities domainIdentity.IdentityRepository,
	userService *domainIdentity.UserService,
	tokenService sessionIssuer,
	subscriptionRepo trialStarter,
) *SignInWithProviderUseCase {
	return &SignInWithProviderUseCase{
		verifier:         verifier,
		users:            users,
		identities:       identities,
		userService:      userService,
		tokenService:     tokenService,
		subscriptionRepo: subscriptionRepo,
	}
}

func (uc *SignInWithProviderUseCase) Execute(ctx context.Context, req SignInWithProviderRequest) (*LoginUserResponse, error) {
	verified, err := uc.verifier.Verify(ctx, strings.ToLower(req.Provider), req.IDToken, req.Nonce)
	if err != nil {
		return nil, err
	}
	user, err := uc.resolveUser(ctx, verified)
	if err != nil {
		return nil, err
	}
	token, refreshToken, err := uc.tokenService.GenerateToken(ctx, user.ID().Value(), user.Email().Value(), user.Role().Value())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	return &LoginUserResponse{
		UserID:       user.ID().Value(),
		Email:        user.Email().Value(),
		Token:        token,
		RefreshToken: refreshToken,
		HasPassword:  user.PasswordHash().IsSet(),
	}, nil
}

func (uc *SignInWithProviderUseCase) resolveUser(ctx context.Context, verified VerifiedSignIn) (*domainIdentity.User, error) {
	existing, err := uc.identities.FindByProviderSubject(ctx, verified.Provider, verified.Subject)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return uc.users.FindByID(ctx, domainIdentity.NewUserID(existing.UserID))
	}
	if verified.Email == "" || !verified.EmailVerified {
		return nil, errors.New("verified email is required")
	}
	user, err := uc.linkOrRegister(ctx, verified)
	if err != nil {
		return nil, err
	}
	if err := uc.identities.Save(ctx, domainIdentity.ExternalIdentity{
		UserID:   user.ID().Value(),
		Provider: verified.Provider,
		Subject:  verified.Subject,
		Email:    verified.Email,
	}); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *SignInWithProviderUseCase) linkOrRegister(ctx context.Context, verified VerifiedSignIn) (*domainIdentity.User, error) {
	email, err := domainIdentity.NewEmail(verified.Email)
	if err != nil {
		return nil, errors.New("invalid email format")
	}
	user, err := uc.users.FindByEmailFolded(ctx, email)
	if err == nil {
		return user, nil
	}
	if !isMissingUser(err) {
		return nil, err
	}
	return uc.registerWithoutPassword(ctx, email)
}

func isMissingUser(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return message == "not found" || message == "record not found"
}

func (uc *SignInWithProviderUseCase) registerWithoutPassword(ctx context.Context, email domainIdentity.Email) (*domainIdentity.User, error) {
	user, err := uc.userService.RegisterUser(ctx, email, domainIdentity.NewPasswordHash(""))
	if err != nil {
		return nil, err
	}
	sub, err := billing.NewTrialSubscription(user.ID().Value(), user.PublicID())
	if err != nil {
		return nil, err
	}
	if err := uc.subscriptionRepo.Save(ctx, sub); err != nil {
		return nil, errors.New("failed to create subscription")
	}
	return user, nil
}

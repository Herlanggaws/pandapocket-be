package identity

import (
	"context"

	"panda-pocket/internal/infrastructure/oauth"
)

type oauthTokenVerifier struct {
	inner *oauth.IDTokenVerifier
}

func NewOAuthTokenVerifier(inner *oauth.IDTokenVerifier) SignInTokenVerifier {
	return oauthTokenVerifier{inner: inner}
}

func (v oauthTokenVerifier) Verify(ctx context.Context, provider, idToken, nonce string) (VerifiedSignIn, error) {
	verified, err := v.inner.Verify(ctx, provider, idToken, nonce)
	if err != nil {
		return VerifiedSignIn{}, err
	}
	return VerifiedSignIn{
		Provider:      verified.Provider,
		Subject:       verified.Subject,
		Email:         verified.Email,
		EmailVerified: verified.EmailVerified,
	}, nil
}

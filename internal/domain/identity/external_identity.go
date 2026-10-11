package identity

import "context"

// ExternalIdentity is a verified Google or Apple subject for one user.
type ExternalIdentity struct {
	UserID   int
	Provider string
	Subject  string
	Email    string
}

// IdentityRepository stores provider subjects separately from the password.
type IdentityRepository interface {
	FindByProviderSubject(ctx context.Context, provider, subject string) (*ExternalIdentity, error)
	Save(ctx context.Context, identity ExternalIdentity) error
}

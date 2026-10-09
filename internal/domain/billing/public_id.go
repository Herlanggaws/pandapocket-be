package billing

import "context"

// PublicIDLookup resolves the random external id for a user.
// It is not a foreign key and does not replace users.id.
type PublicIDLookup interface {
	PublicID(ctx context.Context, userID int) (string, error)
}

package billing

import (
	"context"
	"errors"

	domainBilling "panda-pocket/internal/domain/billing"
)

func lookupPublicID(lookup domainBilling.PublicIDLookup, ctx context.Context, userID int) (string, error) {
	if lookup == nil {
		return "", errors.New("public id lookup is not configured")
	}
	publicID, err := lookup.PublicID(ctx, userID)
	if err != nil {
		return "", err
	}
	if publicID == "" {
		return "", errors.New("public id is required")
	}
	return publicID, nil
}

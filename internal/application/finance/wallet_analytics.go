package finance

import (
	"panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/analytics"
)

// emitWalletCreated records wallet_created only when a wallet was actually created.
// Properties must never include PII (name, amounts, descriptions).
func emitWalletCreated(userID int, wallet *finance.Wallet, created bool, source string) {
	if !created || wallet == nil {
		return
	}
	analytics.Capture(userID, "wallet_created", map[string]interface{}{
		"source":      source,
		"wallet_type": string(wallet.Type()),
	})
}

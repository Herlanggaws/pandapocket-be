package identity

import (
	"crypto/sha256"
	"encoding/hex"
)

// SessionTokenHash is the SHA-256 hex stored for a session JWT.
// An empty token stays empty so a missing credential does not match a row.
func SessionTokenHash(plain string) string {
	if plain == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// IsSessionTokenHash reports whether value is already a stored session hash.
func IsSessionTokenHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

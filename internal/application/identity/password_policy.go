package identity

import "strings"

func passwordRepeatsEmail(password, email string) bool {
	return strings.EqualFold(strings.TrimSpace(password), strings.TrimSpace(email))
}

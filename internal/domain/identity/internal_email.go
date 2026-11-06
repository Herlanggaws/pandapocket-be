package identity

import (
	"os"
	"strings"
)

// InternalEmailDomainsFromEnv reads INTERNAL_EMAIL_DOMAINS (comma-separated).
// Empty or unset means no domain-based internal matching.
func InternalEmailDomainsFromEnv() []string {
	return ParseInternalEmailDomains(os.Getenv("INTERNAL_EMAIL_DOMAINS"))
}

// ParseInternalEmailDomains splits a comma-separated domain list and trims blanks.
func ParseInternalEmailDomains(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	domains := make([]string, 0, len(parts))
	for _, part := range parts {
		domain := strings.ToLower(strings.TrimSpace(part))
		if domain == "" {
			continue
		}
		domains = append(domains, domain)
	}
	return domains
}

// IsInternalEmail reports whether an address should be auto-marked internal.
// Matches when the local part contains "+test" immediately before "@" or
// followed by more local-part characters (e.g. +test@ or +testfoo@), or when
// the domain is listed in domains (case-insensitive).
func IsInternalEmail(email string, domains []string) bool {
	normalized := strings.ToLower(strings.TrimSpace(email))
	at := strings.LastIndex(normalized, "@")
	if at <= 0 || at == len(normalized)-1 {
		return false
	}
	local := normalized[:at]
	domain := normalized[at+1:]

	if strings.Contains(local, "+test") {
		return true
	}
	for _, allowed := range domains {
		if domain == allowed {
			return true
		}
	}
	return false
}

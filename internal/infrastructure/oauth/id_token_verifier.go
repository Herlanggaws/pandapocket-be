package oauth

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
	appleCertsURL  = "https://appleid.apple.com/auth/keys"
	ProviderGoogle = "google"
	ProviderApple  = "apple"
)

type VerifiedIdentity struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
}

type IDTokenVerifier struct {
	googleClientIDs []string
	appleClientIDs  []string
	googleCertsURL  string
	appleCertsURL   string
	client          *http.Client
	mu              sync.Mutex
	keys            map[string]cachedKeySet
}

type cachedKeySet struct {
	byKid   map[string]*rsa.PublicKey
	expires time.Time
}

func NewIDTokenVerifier() *IDTokenVerifier {
	return &IDTokenVerifier{
		googleClientIDs: splitClientIDs(os.Getenv("GOOGLE_OAUTH_CLIENT_IDS")),
		appleClientIDs:  splitClientIDs(os.Getenv("APPLE_OAUTH_CLIENT_IDS")),
		googleCertsURL:  googleCertsURL,
		appleCertsURL:   appleCertsURL,
		client:          &http.Client{Timeout: 5 * time.Second},
		keys:            map[string]cachedKeySet{},
	}
}

func splitClientIDs(raw string) []string {
	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func (v *IDTokenVerifier) Verify(ctx context.Context, provider, idToken, nonce string) (VerifiedIdentity, error) {
	if nonce == "" {
		return VerifiedIdentity{}, errors.New("nonce is required")
	}
	certsURL, clientIDs, err := v.providerConfig(provider)
	if err != nil {
		return VerifiedIdentity{}, err
	}
	claims, err := v.parse(ctx, certsURL, idToken)
	if err != nil {
		return VerifiedIdentity{}, err
	}
	if err := validateAudience(claims, clientIDs); err != nil {
		return VerifiedIdentity{}, err
	}
	if err := validateIssuer(provider, claims); err != nil {
		return VerifiedIdentity{}, err
	}
	if err := validateNonce(provider, claims, nonce); err != nil {
		return VerifiedIdentity{}, err
	}
	return identityFromClaims(provider, claims)
}

func (v *IDTokenVerifier) providerConfig(provider string) (string, []string, error) {
	switch provider {
	case ProviderGoogle:
		if len(v.googleClientIDs) == 0 {
			return "", nil, errors.New("google sign-in is not configured")
		}
		return v.googleCertsURL, v.googleClientIDs, nil
	case ProviderApple:
		if len(v.appleClientIDs) == 0 {
			return "", nil, errors.New("apple sign-in is not configured")
		}
		return v.appleCertsURL, v.appleClientIDs, nil
	default:
		return "", nil, errors.New("unsupported provider")
	}
}

func (v *IDTokenVerifier) parse(ctx context.Context, certsURL, idToken string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		kid, _ := token.Header["kid"].(string)
		return v.publicKey(ctx, certsURL, kid)
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithLeeway(time.Minute))
	if err != nil {
		return nil, errors.New("invalid id token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid id token")
	}
	return claims, nil
}

func (v *IDTokenVerifier) publicKey(ctx context.Context, certsURL, kid string) (*rsa.PublicKey, error) {
	set, err := v.keySet(ctx, certsURL)
	if err != nil {
		return nil, err
	}
	key, ok := set[kid]
	if !ok {
		return nil, errors.New("unknown signing key")
	}
	return key, nil
}

func (v *IDTokenVerifier) keySet(ctx context.Context, certsURL string) (map[string]*rsa.PublicKey, error) {
	v.mu.Lock()
	cached, ok := v.keys[certsURL]
	v.mu.Unlock()
	if ok && time.Now().Before(cached.expires) {
		return cached.byKid, nil
	}
	keys, err := v.fetchKeys(ctx, certsURL)
	if err != nil {
		return nil, err
	}
	v.mu.Lock()
	v.keys[certsURL] = cachedKeySet{byKid: keys, expires: time.Now().Add(time.Hour)}
	v.mu.Unlock()
	return keys, nil
}

func (v *IDTokenVerifier) fetchKeys(ctx context.Context, certsURL string) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, certsURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := v.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to load signing keys")
	}
	var document struct {
		Keys []jsonWebKey `json:"keys"`
	}
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return nil, err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, jwk := range document.Keys {
		publicKey, err := jwk.rsaPublicKey()
		if err != nil || jwk.Kid == "" {
			continue
		}
		keys[jwk.Kid] = publicKey
	}
	if len(keys) == 0 {
		return nil, errors.New("failed to load signing keys")
	}
	return keys, nil
}

type jsonWebKey struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Kty string `json:"kty"`
}

func (key jsonWebKey) rsaPublicKey() (*rsa.PublicKey, error) {
	if key.Kty != "" && key.Kty != "RSA" {
		return nil, errors.New("unsupported key type")
	}
	modulus, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}
	exponent, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: int(new(big.Int).SetBytes(exponent).Int64()),
	}, nil
}

func validateAudience(claims jwt.MapClaims, clientIDs []string) error {
	audiences := audienceValues(claims["aud"])
	for _, audience := range audiences {
		for _, clientID := range clientIDs {
			if audience == clientID {
				return nil
			}
		}
	}
	return errors.New("invalid id token audience")
}

func audienceValues(raw interface{}) []string {
	switch value := raw.(type) {
	case string:
		return []string{value}
	case []interface{}:
		audiences := make([]string, 0, len(value))
		for _, item := range value {
			if audience, ok := item.(string); ok {
				audiences = append(audiences, audience)
			}
		}
		return audiences
	default:
		return nil
	}
}

func validateIssuer(provider string, claims jwt.MapClaims) error {
	issuer, _ := claims["iss"].(string)
	switch provider {
	case ProviderGoogle:
		if issuer == "https://accounts.google.com" || issuer == "accounts.google.com" {
			return nil
		}
	case ProviderApple:
		if issuer == "https://appleid.apple.com" {
			return nil
		}
	}
	return errors.New("invalid id token issuer")
}

func validateNonce(provider string, claims jwt.MapClaims, nonce string) error {
	claim, _ := claims["nonce"].(string)
	expected := nonce
	if provider == ProviderApple {
		sum := sha256.Sum256([]byte(nonce))
		expected = hex.EncodeToString(sum[:])
	}
	if claim == "" || claim != expected {
		return errors.New("invalid id token nonce")
	}
	return nil
}

func identityFromClaims(provider string, claims jwt.MapClaims) (VerifiedIdentity, error) {
	subject, _ := claims["sub"].(string)
	if subject == "" {
		return VerifiedIdentity{}, errors.New("invalid id token subject")
	}
	email, _ := claims["email"].(string)
	email = strings.ToLower(strings.TrimSpace(email))
	verified := emailVerified(claims["email_verified"])
	if provider == ProviderApple && email != "" {
		verified = true
	}
	if email != "" && !verified {
		return VerifiedIdentity{}, errors.New("email is not verified")
	}
	return VerifiedIdentity{
		Provider:      provider,
		Subject:       subject,
		Email:         email,
		EmailVerified: verified,
	}, nil
}

func emailVerified(raw interface{}) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	default:
		return false
	}
}

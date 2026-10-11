package oauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyGoogleIDToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []jsonWebKey{publicJWK("kid-1", &privateKey.PublicKey)},
		})
	}))
	defer server.Close()

	verifier := NewIDTokenVerifier()
	verifier.googleClientIDs = []string{"web-client"}
	verifier.googleCertsURL = server.URL
	verifier.client = server.Client()

	token := signToken(t, privateKey, jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            "web-client",
		"sub":            "google-sub",
		"email":          "Ada@Example.com",
		"email_verified": true,
		"nonce":          "raw-nonce",
		"exp":            time.Now().Add(time.Hour).Unix(),
	})

	verified, err := verifier.Verify(context.Background(), ProviderGoogle, token, "raw-nonce")
	if err != nil {
		t.Fatal(err)
	}
	if verified.Email != "ada@example.com" || verified.Subject != "google-sub" || !verified.EmailVerified {
		t.Fatalf("%+v", verified)
	}
}

func TestVerifyRejectsBadNonce(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []jsonWebKey{publicJWK("kid-1", &privateKey.PublicKey)},
		})
	}))
	defer server.Close()

	verifier := NewIDTokenVerifier()
	verifier.googleClientIDs = []string{"web-client"}
	verifier.googleCertsURL = server.URL
	verifier.client = server.Client()
	token := signToken(t, privateKey, jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            "web-client",
		"sub":            "google-sub",
		"email":          "ada@example.com",
		"email_verified": true,
		"nonce":          "raw-nonce",
		"exp":            time.Now().Add(time.Hour).Unix(),
	})
	if _, err := verifier.Verify(context.Background(), ProviderGoogle, token, "other-nonce"); err == nil {
		t.Fatal("expected nonce failure")
	}
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "kid-1"
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func publicJWK(kid string, key *rsa.PublicKey) jsonWebKey {
	return jsonWebKey{
		Kid: kid,
		Kty: "RSA",
		N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
	}
}

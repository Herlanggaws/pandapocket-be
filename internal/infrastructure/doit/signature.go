package doit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidSignatureHeader = errors.New("invalid PayBridge-Signature header")
	ErrSignatureMismatch      = errors.New("PayBridge-Signature mismatch")
	ErrSignatureExpired       = errors.New("PayBridge-Signature timestamp expired")
)

// VerifyPayBridgeSignature checks HMAC-SHA256(secret, t+"."+rawBody) against header v1.
func VerifyPayBridgeSignature(secret string, header string, rawBody []byte) error {
	if secret == "" {
		return ErrSignatureMismatch
	}
	timestamp, provided, err := parseSignatureHeader(header)
	if err != nil {
		return err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(provided)) {
		return ErrSignatureMismatch
	}
	unixSeconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrInvalidSignatureHeader
	}
	skew := time.Since(time.Unix(unixSeconds, 0))
	if skew > 5*time.Minute || skew < -5*time.Minute {
		return ErrSignatureExpired
	}
	return nil
}

func parseSignatureHeader(header string) (timestamp, v1 string, err error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", "", ErrInvalidSignatureHeader
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}
	if timestamp == "" || v1 == "" {
		return "", "", ErrInvalidSignatureHeader
	}
	return timestamp, v1, nil
}

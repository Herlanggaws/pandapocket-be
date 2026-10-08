package doit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"
	"time"
)

func TestVerifyPayBridgeSignature(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"id":"evt_1","type":"webhook.test"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	v1 := hex.EncodeToString(mac.Sum(nil))
	header := "t=" + ts + ",v1=" + v1

	t.Run("valid", func(t *testing.T) {
		if err := VerifyPayBridgeSignature(secret, header, body); err != nil {
			t.Fatalf("expected valid signature, got %v", err)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		bad := "t=" + ts + ",v1=deadbeef"
		if err := VerifyPayBridgeSignature(secret, bad, body); err != ErrSignatureMismatch {
			t.Fatalf("expected mismatch, got %v", err)
		}
	})

	t.Run("missing header", func(t *testing.T) {
		if err := VerifyPayBridgeSignature(secret, "", body); err != ErrInvalidSignatureHeader {
			t.Fatalf("expected invalid header, got %v", err)
		}
	})

	t.Run("tampered body", func(t *testing.T) {
		if err := VerifyPayBridgeSignature(secret, header, []byte(`{}`)); err != ErrSignatureMismatch {
			t.Fatalf("expected mismatch, got %v", err)
		}
	})

	t.Run("expired timestamp", func(t *testing.T) {
		old := strconv.FormatInt(time.Now().Add(-6*time.Minute).Unix(), 10)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(old + "."))
		mac.Write(body)
		stale := "t=" + old + ",v1=" + hex.EncodeToString(mac.Sum(nil))
		if err := VerifyPayBridgeSignature(secret, stale, body); err != ErrSignatureExpired {
			t.Fatalf("expected expired, got %v", err)
		}
	})
}

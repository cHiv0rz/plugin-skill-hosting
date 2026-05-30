package server

import (
	"strings"
	"testing"

	"marketplace/internal/config"
)

func encTestApp() *App {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return &App{Cfg: config.Config{APITokenKey: key}}
}

func TestEncryptDecryptAPITokenRoundtrip(t *testing.T) {
	a := encTestApp()
	tok := "mkt_" + strings.Repeat("a", 60)
	enc, err := a.encryptAPIToken(tok)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if strings.Contains(enc, tok) {
		t.Error("ciphertext leaks the plaintext token")
	}
	got, err := a.decryptAPIToken(enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != tok {
		t.Errorf("roundtrip = %q, want %q", got, tok)
	}
}

func TestEncryptAPITokenUsesRandomNonce(t *testing.T) {
	a := encTestApp()
	e1, err := a.encryptAPIToken("same-token")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	e2, err := a.encryptAPIToken("same-token")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if e1 == e2 {
		t.Error("identical ciphertext for the same plaintext — nonce is not random")
	}
}

func TestDecryptAPITokenWrongKeyFails(t *testing.T) {
	a := encTestApp()
	enc, err := a.encryptAPIToken("secret-token")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	other := encTestApp()
	other.Cfg.APITokenKey[0] ^= 0xff // a different key
	if _, err := other.decryptAPIToken(enc); err == nil {
		t.Error("decrypt unexpectedly succeeded under a different key")
	}
}

func TestDecryptAPITokenRejectsGarbage(t *testing.T) {
	a := encTestApp()
	if _, err := a.decryptAPIToken("!!!not-base64!!!"); err == nil {
		t.Error("decrypt accepted non-base64 input")
	}
	if _, err := a.decryptAPIToken("YQ"); err == nil { // valid base64, too short for a nonce
		t.Error("decrypt accepted a too-short ciphertext")
	}
}

package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeriveKeyFromSecret(t *testing.T) {
	secret := []byte("test secret")
	salt := []byte("test salt 1234")

	key1 := DeriveKeyFromSecret(secret, salt)

	// Key should be 32 bytes
	if len(key1) != 32 {
		t.Errorf("DeriveKeyFromSecret() length = %d, want 32", len(key1))
	}

	// Same inputs should produce same key
	key2 := DeriveKeyFromSecret(secret, salt)
	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKeyFromSecret() not deterministic")
	}

	// Different salt should produce different key
	differentSalt := []byte("different salt!")
	key3 := DeriveKeyFromSecret(secret, differentSalt)
	if bytes.Equal(key1, key3) {
		t.Error("DeriveKeyFromSecret() same key for different salt")
	}

	// Different secret should produce different key
	differentSecret := []byte("different secret")
	key4 := DeriveKeyFromSecret(differentSecret, salt)
	if bytes.Equal(key1, key4) {
		t.Error("DeriveKeyFromSecret() same key for different secret")
	}
}

func TestDeriveKeyFromPassphrase(t *testing.T) {
	passphrase := "abandon ability able about above"
	salt := []byte("test salt")

	key := DeriveKeyFromPassphrase(passphrase, salt)

	if len(key) != 32 {
		t.Errorf("DeriveKeyFromPassphrase() length = %d, want 32", len(key))
	}

	// Should normalize case
	uppercaseKey := DeriveKeyFromPassphrase(strings.ToUpper(passphrase), salt)
	if !bytes.Equal(key, uppercaseKey) {
		t.Error("DeriveKeyFromPassphrase() not case-insensitive")
	}

	// Should normalize whitespace
	extraSpaces := "  abandon   ability  able   about  above  "
	spacesKey := DeriveKeyFromPassphrase(extraSpaces, salt)
	if !bytes.Equal(key, spacesKey) {
		t.Error("DeriveKeyFromPassphrase() not whitespace-normalized")
	}
}

func TestGenerateRecoveryPhrase(t *testing.T) {
	phrase1, err := GenerateRecoveryPhrase()
	if err != nil {
		t.Fatalf("GenerateRecoveryPhrase() error = %v", err)
	}

	words := strings.Fields(phrase1)
	if len(words) != 12 {
		t.Errorf("GenerateRecoveryPhrase() word count = %d, want 12", len(words))
	}

	// All words should be lowercase
	for _, word := range words {
		if word != strings.ToLower(word) {
			t.Errorf("GenerateRecoveryPhrase() word not lowercase: %s", word)
		}
	}

	// Generate another and ensure they're different
	phrase2, err := GenerateRecoveryPhrase()
	if err != nil {
		t.Fatalf("GenerateRecoveryPhrase() error = %v", err)
	}

	if phrase1 == phrase2 {
		t.Error("GenerateRecoveryPhrase() generated identical phrases")
	}
}

func TestHashRecoveryPhrase(t *testing.T) {
	phrase := "abandon ability able about above absent absorb abstract absurd abuse access accident"

	hash1 := HashRecoveryPhrase(phrase)
	hash2 := HashRecoveryPhrase(phrase)

	// Hash should be 32 bytes (SHA-256)
	if len(hash1) != 32 {
		t.Errorf("HashRecoveryPhrase() length = %d, want 32", len(hash1))
	}

	// Same phrase should produce same hash
	if !bytes.Equal(hash1, hash2) {
		t.Error("HashRecoveryPhrase() not deterministic")
	}

	// Case insensitive
	uppercaseHash := HashRecoveryPhrase(strings.ToUpper(phrase))
	if !bytes.Equal(hash1, uppercaseHash) {
		t.Error("HashRecoveryPhrase() not case-insensitive")
	}
}

func TestVerifyRecoveryPhrase(t *testing.T) {
	phrase := "abandon ability able about above absent absorb abstract absurd abuse access accident"
	hash := HashRecoveryPhrase(phrase)

	tests := []struct {
		name      string
		phrase    string
		wantMatch bool
	}{
		{
			name:      "exact match",
			phrase:    phrase,
			wantMatch: true,
		},
		{
			name:      "uppercase match",
			phrase:    strings.ToUpper(phrase),
			wantMatch: true,
		},
		{
			name:      "with extra spaces",
			phrase:    "  " + phrase + "  ",
			wantMatch: true,
		},
		{
			name:      "wrong phrase",
			phrase:    "wrong phrase here",
			wantMatch: false,
		},
		{
			name:      "empty phrase",
			phrase:    "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyRecoveryPhrase(tt.phrase, hash); got != tt.wantMatch {
				t.Errorf("VerifyRecoveryPhrase() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestEncryptDecryptDEK(t *testing.T) {
	dek, _ := GenerateDEK()
	kek := make([]byte, 32) // Zero key for testing

	encryptedDEK, err := EncryptDEK(dek, kek)
	if err != nil {
		t.Fatalf("EncryptDEK() error = %v", err)
	}

	decryptedDEK, err := DecryptDEK(encryptedDEK, kek)
	if err != nil {
		t.Fatalf("DecryptDEK() error = %v", err)
	}

	if !bytes.Equal(dek, decryptedDEK) {
		t.Error("DecryptDEK() did not return original DEK")
	}
}

func TestGenerateRandomHex(t *testing.T) {
	hex1, err := GenerateRandomHex(16)
	if err != nil {
		t.Fatalf("GenerateRandomHex() error = %v", err)
	}

	// 16 bytes = 32 hex characters
	if len(hex1) != 32 {
		t.Errorf("GenerateRandomHex(16) length = %d, want 32", len(hex1))
	}

	// Should be valid hex
	for _, c := range hex1 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("GenerateRandomHex() invalid character: %c", c)
		}
	}

	// Generate another and ensure they're different
	hex2, _ := GenerateRandomHex(16)
	if hex1 == hex2 {
		t.Error("GenerateRandomHex() generated identical values")
	}
}

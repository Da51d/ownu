package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
		wantErr   bool
	}{
		{
			name:      "valid encryption and decryption",
			plaintext: []byte("Hello, World!"),
			key:       make([]byte, 32),
			wantErr:   false,
		},
		{
			name:      "empty plaintext",
			plaintext: []byte{},
			key:       make([]byte, 32),
			wantErr:   false,
		},
		{
			name:      "large plaintext",
			plaintext: bytes.Repeat([]byte("a"), 10000),
			key:       make([]byte, 32),
			wantErr:   false,
		},
		{
			name:      "invalid key length - too short",
			plaintext: []byte("test"),
			key:       make([]byte, 16),
			wantErr:   true,
		},
		{
			name:      "invalid key length - too long",
			plaintext: []byte("test"),
			key:       make([]byte, 64),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := Encrypt(tt.plaintext, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Ciphertext should be different from plaintext
			if bytes.Equal(ciphertext, tt.plaintext) && len(tt.plaintext) > 0 {
				t.Error("Encrypt() ciphertext equals plaintext")
			}

			// Decrypt and verify
			decrypted, err := Decrypt(ciphertext, tt.key)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{
			name:       "empty ciphertext",
			ciphertext: []byte{},
		},
		{
			name:       "ciphertext too short",
			ciphertext: []byte{1, 2, 3},
		},
		{
			name:       "corrupted ciphertext",
			ciphertext: bytes.Repeat([]byte{0xFF}, 50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, key)
			if err == nil {
				t.Error("Decrypt() expected error for invalid ciphertext")
			}
		})
	}
}

func TestEncryptDecryptString(t *testing.T) {
	key := make([]byte, 32)
	plaintext := "Test string with unicode: 你好世界 🌍"

	encrypted, err := EncryptString(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}

	decrypted, err := DecryptString(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptString() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("DecryptString() = %v, want %v", decrypted, plaintext)
	}
}

func TestGenerateDEK(t *testing.T) {
	dek1, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK() error = %v", err)
	}

	if len(dek1) != 32 {
		t.Errorf("GenerateDEK() length = %d, want 32", len(dek1))
	}

	// Generate another DEK and ensure they're different
	dek2, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK() error = %v", err)
	}

	if bytes.Equal(dek1, dek2) {
		t.Error("GenerateDEK() generated identical keys")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}

	if len(salt1) != 16 {
		t.Errorf("GenerateSalt() length = %d, want 16", len(salt1))
	}

	// Generate another salt and ensure they're different
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("GenerateSalt() generated identical salts")
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("same plaintext")

	// Encrypt the same plaintext twice
	ct1, _ := Encrypt(plaintext, key)
	ct2, _ := Encrypt(plaintext, key)

	// Ciphertexts should be different due to random nonce
	if bytes.Equal(ct1, ct2) {
		t.Error("Encrypt() produced identical ciphertexts for same plaintext (missing randomness)")
	}

	// But both should decrypt to the same plaintext
	pt1, _ := Decrypt(ct1, key)
	pt2, _ := Decrypt(ct2, key)

	if !bytes.Equal(pt1, pt2) || !bytes.Equal(pt1, plaintext) {
		t.Error("Different ciphertexts decrypted to different plaintexts")
	}
}

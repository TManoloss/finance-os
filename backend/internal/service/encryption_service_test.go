package service

import (
	"testing"
)

func TestEncryptionService(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef" // 32 bytes
	s, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}

	tests := []struct {
		name      string
		plainText string
	}{
		{
			name:      "Short string",
			plainText: "hello",
		},
		{
			name:      "Long string",
			plainText: "This is a much longer string to test if the encryption and decryption work correctly for larger payloads. It should handle various characters and lengths without any issues.",
		},
		{
			name:      "Empty string",
			plainText: "",
		},
		{
			name:      "Special characters",
			plainText: "!@#$%^&*()_+{}|:<>?~`-=[]\\;',./",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := s.Encrypt(tt.plainText)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			if encrypted == tt.plainText && tt.plainText != "" {
				t.Errorf("Encrypt() returned plaintext, expected ciphertext")
			}

			decrypted, err := s.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			if decrypted != tt.plainText {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plainText)
			}
		})
	}
}

func TestEncryptionService_InvalidKey(t *testing.T) {
	_, err := NewEncryptionService("invalid")
	if err == nil {
		t.Error("expected error for invalid key length, got nil")
	}
}

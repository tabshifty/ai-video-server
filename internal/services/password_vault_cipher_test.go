package services

import "testing"

func TestPasswordVaultCipherRoundTrip(t *testing.T) {
	t.Parallel()

	cipher, err := NewPasswordVaultCipher("test-password-vault-key")
	if err != nil {
		t.Fatalf("NewPasswordVaultCipher returned error: %v", err)
	}

	first, err := cipher.Encrypt("secret-password")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	second, err := cipher.Encrypt("secret-password")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if first == second {
		t.Fatalf("expected randomized ciphertext, got identical payloads")
	}

	got, err := cipher.Decrypt(first)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if got != "secret-password" {
		t.Fatalf("Decrypt = %q, want %q", got, "secret-password")
	}
}

func TestPasswordVaultCipherRejectsWrongKey(t *testing.T) {
	t.Parallel()

	one, err := NewPasswordVaultCipher("one-key")
	if err != nil {
		t.Fatalf("NewPasswordVaultCipher one returned error: %v", err)
	}
	two, err := NewPasswordVaultCipher("two-key")
	if err != nil {
		t.Fatalf("NewPasswordVaultCipher two returned error: %v", err)
	}

	payload, err := one.Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if _, err := two.Decrypt(payload); err == nil {
		t.Fatalf("expected decrypt with wrong key to fail")
	}
}

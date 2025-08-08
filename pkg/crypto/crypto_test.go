package crypto

import (
	"testing"
)

func TestListClassicAlgorithms(t *testing.T) {
	algorithms := ListClassicAlgorithms()
	if len(algorithms) == 0 {
		t.Error("Expected at least one classic algorithm, got none")
	}

	// Check if AES-256-GCM is in the list
	found := false
	for _, algo := range algorithms {
		if algo.Name == "aes256gcm" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find aes256gcm in classic algorithms")
	}
}

func TestListPostQuantumAlgorithms(t *testing.T) {
	algorithms := ListPostQuantumAlgorithms()
	if len(algorithms) == 0 {
		t.Error("Expected at least one post-quantum algorithm, got none")
	}

	// Check if Kyber-768 is in the list
	found := false
	for _, algo := range algorithms {
		if algo.Name == "kyber768" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find kyber768 in post-quantum algorithms")
	}
}

func TestAES256GCM(t *testing.T) {
	// Test data
	data := []byte("This is a test message for AES-256-GCM encryption")

	// Test the algorithm
	result, err := TestAlgorithm("aes256gcm", data)
	if err != nil {
		t.Errorf("Error testing AES-256-GCM: %v", err)
	}

	// Check the result
	if !result.DecryptionSuccessful {
		t.Error("Decryption was not successful")
	}

	if len(result.Encrypted) == 0 {
		t.Error("Encrypted data is empty")
	}

	if result.KeyGenTime.Nanoseconds() <= 0 {
		t.Error("Key generation time should be positive")
	}

	if result.EncryptTime.Nanoseconds() <= 0 {
		t.Error("Encryption time should be positive")
	}

	if result.DecryptTime.Nanoseconds() <= 0 {
		t.Error("Decryption time should be positive")
	}
}

func TestChaCha20Poly1305(t *testing.T) {
	// Test data
	data := []byte("This is a test message for ChaCha20-Poly1305 encryption")

	// Test the algorithm
	result, err := TestAlgorithm("chacha20poly1305", data)
	if err != nil {
		t.Errorf("Error testing ChaCha20-Poly1305: %v", err)
	}

	// Check the result
	if !result.DecryptionSuccessful {
		t.Error("Decryption was not successful")
	}

	if len(result.Encrypted) == 0 {
		t.Error("Encrypted data is empty")
	}
}

func TestKyber768(t *testing.T) {
	// Test data
	data := []byte("This is a test message for Kyber-768 post-quantum encryption")

	// Test the algorithm
	result, err := TestAlgorithm("kyber768", data)
	if err != nil {
		t.Errorf("Error testing Kyber-768: %v", err)
	}

	// Check the result
	if !result.DecryptionSuccessful {
		t.Error("Decryption was not successful")
	}

	if len(result.Encrypted) == 0 {
		t.Error("Encrypted data is empty")
	}
}

func TestHybridKyberAES(t *testing.T) {
	// Test data
	data := []byte("This is a test message for hybrid Kyber-768 + AES-256-GCM encryption")

	// Test the algorithm
	result, err := TestAlgorithm("hybrid-kyber768-aes256gcm", data)
	if err != nil {
		t.Errorf("Error testing hybrid Kyber-768 + AES-256-GCM: %v", err)
	}

	// Check the result
	if !result.DecryptionSuccessful {
		t.Error("Decryption was not successful")
	}

	if len(result.Encrypted) == 0 {
		t.Error("Encrypted data is empty")
	}

	// The hybrid encryption should produce larger ciphertext than either algorithm alone
	aesResult, _ := TestAlgorithm("aes256gcm", data)
	kyberResult, _ := TestAlgorithm("kyber768", data)

	if len(result.Encrypted) <= len(aesResult.Encrypted) || len(result.Encrypted) <= len(kyberResult.Encrypted) {
		t.Error("Hybrid encryption should produce larger ciphertext than either algorithm alone")
	}
}

func TestUnsupportedAlgorithm(t *testing.T) {
	// Test data
	data := []byte("This is a test message")

	// Test an unsupported algorithm
	_, err := TestAlgorithm("unsupported-algorithm", data)
	if err == nil {
		t.Error("Expected error for unsupported algorithm, got nil")
	}
}
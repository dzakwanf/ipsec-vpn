package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cloudflare/circl/kem"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
	"github.com/cloudflare/circl/kem/kyber/kyber1024"
	"github.com/dzakwan/ipsec-vpn/pkg/logger"
	"github.com/spf13/viper"
	"golang.org/x/crypto/chacha20poly1305"
)

// Algorithm represents a cryptographic algorithm
type Algorithm struct {
	Name        string
	Description string
	PostQuantum bool
}

// TestResult represents the result of testing an algorithm
type TestResult struct {
	Algorithm            string
	Encrypted            []byte
	DecryptionSuccessful bool
	KeyGenTime           time.Duration
	EncryptTime          time.Duration
	DecryptTime          time.Duration
}

// ListClassicAlgorithms returns a list of available classic encryption algorithms
func ListClassicAlgorithms() []Algorithm {
	logger.Debug("Listing available classic encryption algorithms")
	return []Algorithm{
		{
			Name:        "aes256gcm",
			Description: "AES-256 in GCM mode - Strong symmetric encryption",
			PostQuantum: false,
		},
		{
			Name:        "chacha20poly1305",
			Description: "ChaCha20-Poly1305 - Fast and secure symmetric encryption",
			PostQuantum: false,
		},
	}
}

// ListPostQuantumAlgorithms returns a list of available post-quantum encryption algorithms
func ListPostQuantumAlgorithms() []Algorithm {
	logger.Debug("Listing available post-quantum encryption algorithms")
	return []Algorithm{
		{
			Name:        "kyber768",
			Description: "Kyber-768 - NIST selected post-quantum key encapsulation mechanism",
			PostQuantum: true,
		},
		{
			Name:        "kyber1024",
			Description: "Kyber-1024 - Higher security level post-quantum key encapsulation mechanism",
			PostQuantum: true,
		},
		{
			Name:        "hybrid-kyber768-aes256gcm",
			Description: "Hybrid Kyber-768 + AES-256-GCM - Post-quantum security with classical fallback",
			PostQuantum: true,
		},
	}
}

// TestAlgorithm tests an encryption algorithm with the given data
func TestAlgorithm(algorithm string, data []byte) (*TestResult, error) {
	logger.Info("Testing encryption algorithm: %s", algorithm)
	result := &TestResult{
		Algorithm: algorithm,
	}

	// Test the algorithm based on its type
	switch algorithm {
	case "aes256gcm":
		logger.Debug("Testing AES-256-GCM algorithm")
		return testAES256GCM(data, result)
	case "chacha20poly1305":
		logger.Debug("Testing ChaCha20-Poly1305 algorithm")
		return testChaCha20Poly1305(data, result)
	case "kyber768":
		logger.Debug("Testing Kyber-768 algorithm")
		return testKyber(kyber768.Scheme(), data, result)
	case "kyber1024":
		logger.Debug("Testing Kyber-1024 algorithm")
		return testKyber(kyber1024.Scheme(), data, result)
	case "hybrid-kyber768-aes256gcm":
		logger.Debug("Testing Hybrid Kyber-768 + AES-256-GCM algorithm")
		return testHybridKyberAES(kyber768.Scheme(), data, result)
	default:
		logger.Error("Unsupported algorithm: %s", algorithm)
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// SetDefaultAlgorithm sets the default encryption algorithm
func SetDefaultAlgorithm(algorithm string, postQuantum bool) error {
	logger.Info("Setting default encryption algorithm to %s (post-quantum: %t)", algorithm, postQuantum)
	// Validate algorithm
	valid := false
	if postQuantum {
		for _, algo := range ListPostQuantumAlgorithms() {
			if algo.Name == algorithm {
				valid = true
				break
			}
		}
	} else {
		for _, algo := range ListClassicAlgorithms() {
			if algo.Name == algorithm {
				valid = true
				break
			}
		}
	}

	if !valid {
		return fmt.Errorf("invalid algorithm: %s", algorithm)
	}

	// Set default algorithm in configuration
	if postQuantum {
		viper.Set("crypto.default_post_quantum", algorithm)
	} else {
		viper.Set("crypto.default_classic", algorithm)
	}

	// Save configuration
	return viper.WriteConfig()
}

// GetDefaultAlgorithm returns the default encryption algorithm
func GetDefaultAlgorithm(postQuantum bool) string {
	if postQuantum {
		defaultAlgo := viper.GetString("crypto.default_post_quantum")
		if defaultAlgo == "" {
			return "kyber768" // Default post-quantum algorithm
		}
		return defaultAlgo
	} else {
		defaultAlgo := viper.GetString("crypto.default_classic")
		if defaultAlgo == "" {
			return "aes256gcm" // Default classic algorithm
		}
		return defaultAlgo
	}
}

// Helper functions

// testAES256GCM tests AES-256-GCM encryption
func testAES256GCM(data []byte, result *TestResult) (*TestResult, error) {
	// Generate key
	startKeyGen := time.Now()
	key := make([]byte, 32) // AES-256 uses a 32-byte key
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	result.KeyGenTime = time.Since(startKeyGen)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt
	startEncrypt := time.Now()
	ciphertext := aead.Seal(nonce, nonce, data, nil)
	result.EncryptTime = time.Since(startEncrypt)
	result.Encrypted = ciphertext

	// Decrypt
	startDecrypt := time.Now()
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonceDecrypt, ciphertextDecrypt := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonceDecrypt, ciphertextDecrypt, nil)
	result.DecryptTime = time.Since(startDecrypt)

	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	result.DecryptionSuccessful = string(plaintext) == string(data)
	return result, nil
}

// testChaCha20Poly1305 tests ChaCha20-Poly1305 encryption
func testChaCha20Poly1305(data []byte, result *TestResult) (*TestResult, error) {
	// Generate key
	startKeyGen := time.Now()
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	result.KeyGenTime = time.Since(startKeyGen)

	// Create cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt
	startEncrypt := time.Now()
	ciphertext := aead.Seal(nonce, nonce, data, nil)
	result.EncryptTime = time.Since(startEncrypt)
	result.Encrypted = ciphertext

	// Decrypt
	startDecrypt := time.Now()
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonceDecrypt, ciphertextDecrypt := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonceDecrypt, ciphertextDecrypt, nil)
	result.DecryptTime = time.Since(startDecrypt)

	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	result.DecryptionSuccessful = string(plaintext) == string(data)
	return result, nil
}

// testKyber tests Kyber post-quantum key encapsulation
func testKyber(scheme kem.Scheme, data []byte, result *TestResult) (*TestResult, error) {
	// Generate key pair
	startKeyGen := time.Now()
	public, private, err := scheme.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	result.KeyGenTime = time.Since(startKeyGen)

	// Encapsulate (encrypt)
	startEncrypt := time.Now()
	ciphertext, sharedSecret, err := scheme.Encapsulate(public)
	if err != nil {
		return nil, err
	}
	result.EncryptTime = time.Since(startEncrypt)

	// Use shared secret to encrypt data with AES-GCM
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encryptedData := aead.Seal(nil, nonce, data, nil)

	// Combine ciphertext, nonce, and encrypted data
	result.Encrypted = append(ciphertext, append(nonce, encryptedData...)...)

	// Decapsulate (decrypt)
	startDecrypt := time.Now()
	// Extract Kyber ciphertext
	kyberCiphertextSize := scheme.CiphertextSize()
	if len(result.Encrypted) < kyberCiphertextSize {
		return nil, errors.New("ciphertext too short")
	}

	kyberCiphertext := result.Encrypted[:kyberCiphertextSize]
	remaining := result.Encrypted[kyberCiphertextSize:]

	// Decapsulate to get shared secret
	decapsulatedSecret, err := scheme.Decapsulate(private, kyberCiphertext)
	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	// Use shared secret to decrypt data
	block, err = aes.NewCipher(decapsulatedSecret)
	if err != nil {
		return nil, err
	}

	aead, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(remaining) < nonceSize {
		return nil, errors.New("remaining ciphertext too short")
	}

	nonceDecrypt, ciphertextDecrypt := remaining[:nonceSize], remaining[nonceSize:]
	plaintext, err := aead.Open(nil, nonceDecrypt, ciphertextDecrypt, nil)
	result.DecryptTime = time.Since(startDecrypt)

	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	result.DecryptionSuccessful = string(plaintext) == string(data)
	return result, nil
}

// testHybridKyberAES tests hybrid Kyber+AES encryption
func testHybridKyberAES(scheme kem.Scheme, data []byte, result *TestResult) (*TestResult, error) {
	// Generate key pair for Kyber
	startKeyGen := time.Now()
	public, private, err := scheme.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	// Generate AES key
	aesKey := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, err
	}
	result.KeyGenTime = time.Since(startKeyGen)

	// Encapsulate with Kyber
	startEncrypt := time.Now()
	ciphertext, sharedSecret, err := scheme.Encapsulate(public)
	if err != nil {
		return nil, err
	}

	// Combine Kyber shared secret with AES key
	hybridKey := make([]byte, len(sharedSecret)+len(aesKey))
	copy(hybridKey, sharedSecret)
	copy(hybridKey[len(sharedSecret):], aesKey)

	// Use hybrid key to encrypt data with AES-GCM
	block, err := aes.NewCipher(hybridKey[:32]) // Use first 32 bytes for AES-256
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encryptedData := aead.Seal(nil, nonce, data, nil)
	result.EncryptTime = time.Since(startEncrypt)

	// Combine ciphertext, AES key (encrypted with Kyber shared secret), nonce, and encrypted data
	// Encrypt AES key with Kyber shared secret
	aesKeyBlock, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, err
	}

	aesKeyAead, err := cipher.NewGCM(aesKeyBlock)
	if err != nil {
		return nil, err
	}

	aesKeyNonce := make([]byte, aesKeyAead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, aesKeyNonce); err != nil {
		return nil, err
	}

	encryptedAesKey := aesKeyAead.Seal(nil, aesKeyNonce, aesKey, nil)

	// Combine all components
	result.Encrypted = append(ciphertext, append(aesKeyNonce, append(encryptedAesKey, append(nonce, encryptedData...)...)...)...)

	// Decapsulate and decrypt
	startDecrypt := time.Now()
	// Extract Kyber ciphertext
	kyberCiphertextSize := scheme.CiphertextSize()
	if len(result.Encrypted) < kyberCiphertextSize {
		return nil, errors.New("ciphertext too short")
	}

	kyberCiphertext := result.Encrypted[:kyberCiphertextSize]
	remaining := result.Encrypted[kyberCiphertextSize:]

	// Decapsulate to get shared secret
	decapsulatedSecret, err := scheme.Decapsulate(private, kyberCiphertext)
	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	// Extract AES key nonce and encrypted AES key
	aesKeyBlock, err = aes.NewCipher(decapsulatedSecret)
	if err != nil {
		return nil, err
	}

	aesKeyAead, err = cipher.NewGCM(aesKeyBlock)
	if err != nil {
		return nil, err
	}

	aesKeyNonceSize := aesKeyAead.NonceSize()
	if len(remaining) < aesKeyNonceSize {
		return nil, errors.New("remaining ciphertext too short")
	}

	aesKeyNonceDecrypt, remaining := remaining[:aesKeyNonceSize], remaining[aesKeyNonceSize:]

	// Determine the size of the encrypted AES key (GCM adds a 16-byte tag)
	encryptedAesKeySize := 32 + 16 // AES-256 key + GCM tag
	if len(remaining) < encryptedAesKeySize {
		return nil, errors.New("remaining ciphertext too short for AES key")
	}

	encryptedAesKeyDecrypt, remaining := remaining[:encryptedAesKeySize], remaining[encryptedAesKeySize:]

	// Decrypt the AES key
	decryptedAesKey, err := aesKeyAead.Open(nil, aesKeyNonceDecrypt, encryptedAesKeyDecrypt, nil)
	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	// Combine Kyber shared secret with decrypted AES key
	hybridKeyDecrypt := make([]byte, len(decapsulatedSecret)+len(decryptedAesKey))
	copy(hybridKeyDecrypt, decapsulatedSecret)
	copy(hybridKeyDecrypt[len(decapsulatedSecret):], decryptedAesKey)

	// Use hybrid key to decrypt data
	blockDecrypt, err := aes.NewCipher(hybridKeyDecrypt[:32]) // Use first 32 bytes for AES-256
	if err != nil {
		return nil, err
	}

	aeadDecrypt, err := cipher.NewGCM(blockDecrypt)
	if err != nil {
		return nil, err
	}

	nonceSize := aeadDecrypt.NonceSize()
	if len(remaining) < nonceSize {
		return nil, errors.New("remaining ciphertext too short for data nonce")
	}

	nonceDecrypt, ciphertextDecrypt := remaining[:nonceSize], remaining[nonceSize:]
	plaintext, err := aeadDecrypt.Open(nil, nonceDecrypt, ciphertextDecrypt, nil)
	result.DecryptTime = time.Since(startDecrypt)

	if err != nil {
		result.DecryptionSuccessful = false
		return result, nil
	}

	result.DecryptionSuccessful = string(plaintext) == string(data)
	return result, nil
}
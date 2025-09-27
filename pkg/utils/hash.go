package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type HashConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Default configuration for Argon2id - High security settings
var defaultHashConfig = HashConfig{
	Memory:      128 * 1024, // 128 MB (เพิ่มเป็น 2 เท่า)
	Iterations:  4,          // เพิ่มรอบการคำนวณ
	Parallelism: 4,          // ใช้ CPU มากขึ้น
	SaltLength:  32,         // เพิ่ม salt เป็น 32 bytes
	KeyLength:   64,         // เพิ่ม output เป็น 64 bytes
}

// HashPassword creates a hash of the password using Argon2id
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, defaultHashConfig.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Generate the hash
	hash := argon2.IDKey([]byte(password), salt, defaultHashConfig.Iterations,
		defaultHashConfig.Memory, defaultHashConfig.Parallelism, defaultHashConfig.KeyLength)

	// Encode salt and hash to base64
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, defaultHashConfig.Memory, defaultHashConfig.Iterations,
		defaultHashConfig.Parallelism, saltB64, hashB64)

	return encodedHash, nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}

	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version of argon2")
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Generate hash with provided password and extracted parameters
	otherHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	// Use constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

// CheckPasswordHash is an alias for VerifyPassword for consistency
func CheckPasswordHash(password, hash string) bool {
	match, err := VerifyPassword(password, hash)
	if err != nil {
		return false
	}
	return match
}

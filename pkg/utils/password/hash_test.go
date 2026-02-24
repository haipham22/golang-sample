package password

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "SecurePassword123!",
			expectError: false,
		},
		{
			name:        "short password",
			password:    "pass",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: false,
		},
		{
			name:        "long password (within 72 byte limit)",
			password:    "ThisIsALongPasswordThatIsStillWithin72BytesLimit!",
			expectError: false,
		},
		{
			name:        "password with special chars",
			password:    "P@ssw0rd!#$%^&*()",
			expectError: false,
		},
		{
			name:        "password exceeding 72 bytes (bcrypt limit)",
			password:    string(make([]byte, 100)), // 100 bytes, exceeds bcrypt limit
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && hash == "" {
				t.Errorf("expected hash to be non-empty")
			}

			if !tt.expectError && hash == tt.password {
				t.Errorf("hash should not equal plaintext password")
			}

			// Verify hash is bcrypt format (starts with $2a$, $2b$, or $2y$)
			if !tt.expectError && hash != "" {
				if len(hash) < 60 {
					t.Errorf("bcrypt hash should be at least 60 characters, got %d", len(hash))
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// Setup: Create a hash for testing
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to create test hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "WrongPassword123!",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "password with different case",
			password: "testpassword123!",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: password,
			hash:     "invalid-hash",
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPasswordHash(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	password := "ConsistentPassword123!"

	// Hash the same password twice using MinCost for faster tests
	hash1, err1 := HashPasswordWithCost(password, bcrypt.MinCost)
	hash2, err2 := HashPasswordWithCost(password, bcrypt.MinCost)

	if err1 != nil || err2 != nil {
		t.Fatalf("failed to generate hashes: err1=%v, err2=%v", err1, err2)
	}

	// Hashes should be different (bcrypt uses salt)
	if hash1 == hash2 {
		t.Errorf("hashes should be different due to salt, got same hash: %s", hash1)
	}

	// But both should verify correctly
	if !CheckPasswordHash(password, hash1) {
		t.Errorf("hash1 should verify correctly")
	}
	if !CheckPasswordHash(password, hash2) {
		t.Errorf("hash2 should verify correctly")
	}
}

func TestCheckPasswordHashWithMultipleHashes(t *testing.T) {
	password := "MultiHashTest123!"

	// Create multiple hashes using MinCost for faster tests
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		hash, err := HashPasswordWithCost(password, bcrypt.MinCost)
		if err != nil {
			t.Fatalf("failed to create hash %d: %v", i, err)
		}
		hashes[i] = hash
	}

	// All hashes should verify the same password
	for i, hash := range hashes {
		if !CheckPasswordHash(password, hash) {
			t.Errorf("hash %d failed to verify password", i)
		}
	}

	// Wrong password should fail all hashes
	wrongPassword := "WrongPassword"
	for i, hash := range hashes {
		if CheckPasswordHash(wrongPassword, hash) {
			t.Errorf("hash %d incorrectly verified wrong password", i)
		}
	}
}

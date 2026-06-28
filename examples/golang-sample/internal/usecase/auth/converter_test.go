package auth

import (
	"testing"
	"time"

	"github.com/haipham22/golang-sample/internal/domain"
	"github.com/haipham22/golang-sample/internal/schemas"
)

func TestModelToSchemaUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *domain.User
		expected *schemas.User
	}{
		{
			name: "valid conversion",
			input: &domain.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: &schemas.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modelToSchemaUser(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected non-nil result")
			}

			if result.ID != tt.expected.ID {
				t.Errorf("ID mismatch: got %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("Username mismatch: got %s, want %s", result.Username, tt.expected.Username)
			}
			if result.Email != tt.expected.Email {
				t.Errorf("Email mismatch: got %s, want %s", result.Email, tt.expected.Email)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("CreatedAt mismatch: got %v, want %v", result.CreatedAt, tt.expected.CreatedAt)
			}
			if !result.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("UpdatedAt mismatch: got %v, want %v", result.UpdatedAt, tt.expected.UpdatedAt)
			}
		})
	}
}

func TestSchemaToModelUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *schemas.User
		expected *domain.User
	}{
		{
			name: "valid conversion",
			input: &schemas.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: &domain.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schemaToModelUser(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected non-nil result")
			}

			if result.ID != tt.expected.ID {
				t.Errorf("ID mismatch: got %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("Username mismatch: got %s, want %s", result.Username, tt.expected.Username)
			}
			if result.Email != tt.expected.Email {
				t.Errorf("Email mismatch: got %s, want %s", result.Email, tt.expected.Email)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("CreatedAt mismatch: got %v, want %v", result.CreatedAt, tt.expected.CreatedAt)
			}
			if !result.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("UpdatedAt mismatch: got %v, want %v", result.UpdatedAt, tt.expected.UpdatedAt)
			}
		})
	}
}

func TestSchemaModelRoundTrip(t *testing.T) {
	now := time.Now()

	original := &schemas.User{
		ID:        123,
		Username:  "roundtrip",
		Email:     "roundtrip@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Schema -> Model -> Schema
	modelUser := schemaToModelUser(original)
	backToSchema := modelToSchemaUser(modelUser)

	if backToSchema.ID != original.ID {
		t.Errorf("ID not preserved: got %d, want %d", backToSchema.ID, original.ID)
	}
	if backToSchema.Username != original.Username {
		t.Errorf("Username not preserved: got %s, want %s", backToSchema.Username, original.Username)
	}
	if backToSchema.Email != original.Email {
		t.Errorf("Email not preserved: got %s, want %s", backToSchema.Email, original.Email)
	}
}

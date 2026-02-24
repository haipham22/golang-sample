package user

import (
	"testing"
	"time"

	"golang-sample/internal/model"
	"golang-sample/internal/orm"
)

func TestOrmToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *orm.User
		expected *model.User
	}{
		{
			name: "valid conversion",
			input: &orm.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: &model.User{
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
			result := ormToModel(tt.input)

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

func TestModelToORM(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *model.User
		expected *orm.User
	}{
		{
			name: "valid conversion",
			input: &model.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: &orm.User{
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
			result := modelToORM(tt.input)

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

func TestOrmSliceToModelSlice(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    []*orm.User
		expected []*model.User
	}{
		{
			name: "valid slice conversion",
			input: []*orm.User{
				{
					ID:        1,
					Username:  "user1",
					Email:     "user1@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        2,
					Username:  "user2",
					Email:     "user2@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: []*model.User{
				{
					ID:        1,
					Username:  "user1",
					Email:     "user1@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        2,
					Username:  "user2",
					Email:     "user2@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []*orm.User{},
			expected: []*model.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ormSliceToModelSlice(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("user[%d].ID mismatch: got %d, want %d", i, result[i].ID, tt.expected[i].ID)
				}
				if result[i].Username != tt.expected[i].Username {
					t.Errorf("user[%d].Username mismatch: got %s, want %s", i, result[i].Username, tt.expected[i].Username)
				}
				if result[i].Email != tt.expected[i].Email {
					t.Errorf("user[%d].Email mismatch: got %s, want %s", i, result[i].Email, tt.expected[i].Email)
				}
			}
		})
	}
}

func TestConversionRoundTrip(t *testing.T) {
	now := time.Now()

	original := &model.User{
		ID:        123,
		Username:  "roundtrip",
		Email:     "roundtrip@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Model -> ORM -> Model
	ormUser := modelToORM(original)
	backToModel := ormToModel(ormUser)

	if backToModel.ID != original.ID {
		t.Errorf("ID not preserved: got %d, want %d", backToModel.ID, original.ID)
	}
	if backToModel.Username != original.Username {
		t.Errorf("Username not preserved: got %s, want %s", backToModel.Username, original.Username)
	}
	if backToModel.Email != original.Email {
		t.Errorf("Email not preserved: got %s, want %s", backToModel.Email, original.Email)
	}
}

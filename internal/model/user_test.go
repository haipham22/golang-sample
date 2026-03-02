package model

import (
	"testing"
	"time"
)

func TestUser_ValidationEdgeCases(t *testing.T) {
	t.Run("exactly 3 character username is valid", func(t *testing.T) {
		user := &User{
			Username: "abc",
			Email:    "test@example.com",
		}
		err := user.Validate()
		if err != nil {
			t.Errorf("3 character username should be valid, got error: %v", err)
		}
	})

	t.Run("exactly 50 character username is valid", func(t *testing.T) {
		username := ""
		for i := 0; i < 50; i++ {
			username += string(rune('a' + i%26))
		}
		user := &User{
			Username: username,
			Email:    "test@example.com",
		}
		err := user.Validate()
		if err != nil {
			t.Errorf("50 character username should be valid, got error: %v", err)
		}
	})

	t.Run("51 character username is too long", func(t *testing.T) {
		user := &User{
			Username: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Email:    "test@example.com",
		}
		err := user.Validate()
		if err != ErrUsernameTooLong {
			t.Errorf("51 character username should be too long, got error: %v", err)
		}
	})

	t.Run("2 character username is too short", func(t *testing.T) {
		user := &User{
			Username: "ab",
			Email:    "test@example.com",
		}
		err := user.Validate()
		if err != ErrUsernameTooShort {
			t.Errorf("2 character username should be too short, got error: %v", err)
		}
	})

	t.Run("empty username is required", func(t *testing.T) {
		user := &User{
			Username: "",
			Email:    "test@example.com",
		}
		err := user.Validate()
		if err != ErrUsernameRequired {
			t.Errorf("empty username should trigger required error, got error: %v", err)
		}
	})

	t.Run("empty email is required", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Email:    "",
		}
		err := user.Validate()
		if err != ErrEmailRequired {
			t.Errorf("empty email should trigger required error, got error: %v", err)
		}
	})

	t.Run("both empty triggers username error first", func(t *testing.T) {
		user := &User{
			Username: "",
			Email:    "",
		}
		err := user.Validate()
		if err != ErrUsernameRequired {
			t.Errorf("both empty should trigger username required error first, got error: %v", err)
		}
	})
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name: "valid user",
			user: &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantErr: nil,
		},
		{
			name: "missing username",
			user: &User{
				ID:    1,
				Email: "test@example.com",
			},
			wantErr: ErrUsernameRequired,
		},
		{
			name: "missing email",
			user: &User{
				ID:       1,
				Username: "testuser",
			},
			wantErr: ErrEmailRequired,
		},
		{
			name: "username too short",
			user: &User{
				ID:       1,
				Username: "ab",
				Email:    "test@example.com",
			},
			wantErr: ErrUsernameTooShort,
		},
		{
			name: "username too long",
			user: &User{
				ID:       1,
				Username: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Email:    "test@example.com",
			},
			wantErr: ErrUsernameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if err != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{
			name: "can login with valid data",
			user: &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			want: true,
		},
		{
			name: "cannot login without username",
			user: &User{
				ID:    1,
				Email: "test@example.com",
			},
			want: false,
		},
		{
			name: "cannot login without email",
			user: &User{
				ID:       1,
				Username: "testuser",
			},
			want: false,
		},
		{
			name: "cannot login with empty fields",
			user: &User{
				ID: 1,
			},
			want: false,
		},
		{
			name: "nil user cannot login",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user == nil {
				if got := tt.user.CanLogin(); got != tt.want {
					t.Errorf("User.CanLogin() = %v, want %v", got, tt.want)
				}
				return
			}
			if got := tt.user.CanLogin(); got != tt.want {
				t.Errorf("User.CanLogin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsNew(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{
			name: "user with ID is not new",
			user: &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			want: false,
		},
		{
			name: "user without ID is new",
			user: &User{
				Username: "testuser",
				Email:    "test@example.com",
			},
			want: true,
		},
		{
			name: "user with ID 0 is new",
			user: &User{
				ID:       0,
				Username: "testuser",
				Email:    "test@example.com",
			},
			want: true,
		},
		{
			name: "nil user is new",
			user: nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user == nil {
				if got := tt.user.IsNew(); got != tt.want {
					t.Errorf("User.IsNew() = %v, want %v", got, tt.want)
				}
				return
			}
			if got := tt.user.IsNew(); got != tt.want {
				t.Errorf("User.IsNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsEqual(t *testing.T) {
	tests := []struct {
		name  string
		user  *User
		other *User
		want  bool
	}{
		{
			name: "same users are equal",
			user: &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			other: &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			want: true,
		},
		{
			name: "different IDs are not equal",
			user: &User{
				ID:       1,
				Username: "testuser",
			},
			other: &User{
				ID:       2,
				Username: "testuser",
			},
			want: false,
		},
		{
			name: "different usernames are not equal",
			user: &User{
				ID:       1,
				Username: "user1",
			},
			other: &User{
				ID:       1,
				Username: "user2",
			},
			want: false,
		},
		{
			name:  "nil users are not equal",
			user:  nil,
			other: nil,
			want:  false,
		},
		{
			name: "one nil user is not equal",
			user: &User{
				ID:       1,
				Username: "testuser",
			},
			other: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.IsEqual(tt.other); got != tt.want {
				t.Errorf("User.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Clone(t *testing.T) {
	now := time.Now()
	original := &User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("clone creates deep copy", func(t *testing.T) {
		cloned := original.Clone()

		if cloned == nil {
			t.Fatal("Clone() returned nil")
		}

		if cloned.ID != original.ID {
			t.Errorf("Clone() ID = %v, want %v", cloned.ID, original.ID)
		}
		if cloned.Username != original.Username {
			t.Errorf("Clone() Username = %v, want %v", cloned.Username, original.Username)
		}
		if cloned.Email != original.Email {
			t.Errorf("Clone() Email = %v, want %v", cloned.Email, original.Email)
		}
		if !cloned.CreatedAt.Equal(original.CreatedAt) {
			t.Errorf("Clone() CreatedAt = %v, want %v", cloned.CreatedAt, original.CreatedAt)
		}
		if !cloned.UpdatedAt.Equal(original.UpdatedAt) {
			t.Errorf("Clone() UpdatedAt = %v, want %v", cloned.UpdatedAt, original.UpdatedAt)
		}
	})

	t.Run("clone is independent from original", func(t *testing.T) {
		cloned := original.Clone()
		cloned.Username = "modified"
		cloned.ID = 999

		if original.Username == "modified" {
			t.Error("Modifying clone affected original Username")
		}
		if original.ID == 999 {
			t.Error("Modifying clone affected original ID")
		}
	})

	t.Run("clone of nil returns nil", func(t *testing.T) {
		var nilUser *User
		cloned := nilUser.Clone()
		if cloned != nil {
			t.Errorf("Clone() of nil = %v, want nil", cloned)
		}
	})
}

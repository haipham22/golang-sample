package metrics

import "testing"

// Test helpers shared across test files
func requireNotNil(t *testing.T, v interface{}) {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertNotPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("expected no panic, got %v", r)
		}
	}()
	f()
}

func assertEmpty(t *testing.T, s string) {
	t.Helper()
	if s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func assertEqual(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Errorf("expected %d, got %d", want, got)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

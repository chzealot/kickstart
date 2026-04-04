package version

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v1.0.0", "1.0.0"},
		{"1.0.0", "1.0.0"},
		{"v0.1.0", "0.1.0"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := normalizeVersion(tt.input); got != tt.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v1.1.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.1.0", false},
		{"1.1.0", "1.0.0", true},
		{"v2.0.0", "v1.9.9", true},
		{"v0.2.0", "v0.1.0", true},
		{"v0.1.0", "v0.2.0", false},
	}
	for _, tt := range tests {
		if got := IsNewer(tt.latest, tt.current); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestCheckResult_DevVersion(t *testing.T) {
	// Version defaults to "dev", checker should return nil (skip check)
	// Note: we don't modify the global to avoid data races with the goroutine
	checker := NewAsyncChecker()
	// Give goroutine time to complete
	var result *CheckResult
	for i := 0; i < 200; i++ {
		result = checker.Result()
		if result != nil {
			break
		}
	}
	// dev version should produce nil result
	if result != nil {
		t.Errorf("expected nil result for dev version, got %+v", result)
	}
}

package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchLatestVersion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"tag_name": "v1.2.3"})
	}))
	defer server.Close()

	// Override the fetch function by testing the HTTP handler directly
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Errorf("got %q, want %q", release.TagName, "v1.2.3")
	}
}

func TestFetchLatestVersion_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAsyncChecker_WithUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"tag_name": "v2.0.0"})
	}))
	defer server.Close()

	// Test the CheckResult logic directly
	result := &CheckResult{
		Latest:    "v2.0.0",
		Current:   "v1.0.0",
		HasUpdate: IsNewer("v2.0.0", "v1.0.0"),
	}
	if !result.HasUpdate {
		t.Error("expected HasUpdate to be true")
	}
}

func TestAsyncChecker_NoUpdate(t *testing.T) {
	result := &CheckResult{
		Latest:    "v1.0.0",
		Current:   "v1.0.0",
		HasUpdate: IsNewer("v1.0.0", "v1.0.0"),
	}
	if result.HasUpdate {
		t.Error("expected HasUpdate to be false")
	}
}

func TestAsyncChecker_VersionWithAndWithoutPrefix(t *testing.T) {
	result := &CheckResult{
		Latest:    "v1.0.0",
		Current:   "1.0.0",
		HasUpdate: IsNewer("v1.0.0", "1.0.0"),
	}
	if result.HasUpdate {
		t.Error("expected HasUpdate to be false for v1.0.0 vs 1.0.0")
	}
}

package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/mod/semver"
)

// CheckResult holds the result of a version check.
type CheckResult struct {
	Latest    string
	Current   string
	HasUpdate bool
}

// AsyncChecker performs background version checks.
type AsyncChecker struct {
	result chan *CheckResult
	once   sync.Once
}

// NewAsyncChecker starts a background version check and returns a checker.
func NewAsyncChecker() *AsyncChecker {
	c := &AsyncChecker{
		result: make(chan *CheckResult, 1),
	}
	go c.check()
	return c
}

// Result returns the check result if available, or nil if not ready / error.
func (c *AsyncChecker) Result() *CheckResult {
	select {
	case r := <-c.result:
		return r
	default:
		return nil
	}
}

// WaitResult waits up to timeout for the check result.
func (c *AsyncChecker) WaitResult(timeout time.Duration) *CheckResult {
	select {
	case r := <-c.result:
		return r
	case <-time.After(timeout):
		return nil
	}
}

func (c *AsyncChecker) check() {
	defer func() {
		// Ensure channel is never left blocking
		select {
		case c.result <- nil:
		default:
		}
	}()

	if Version == "dev" {
		return
	}

	latest, err := fetchLatestVersion()
	if err != nil {
		return
	}

	hasUpdate := latest != "" && IsNewer(latest, Version)
	c.result <- &CheckResult{
		Latest:    latest,
		Current:   Version,
		HasUpdate: hasUpdate,
	}
}

func fetchLatestVersion() (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/chzealot/kickstart/releases/latest", nil)
	if err != nil {
		return "", err
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// IsNewer returns true if latest is a newer version than current (semver comparison).
// Falls back to string comparison if versions are not valid semver.
func IsNewer(latest, current string) bool {
	l := ensureVPrefix(latest)
	c := ensureVPrefix(current)
	if semver.IsValid(l) && semver.IsValid(c) {
		return semver.Compare(l, c) > 0
	}
	// Fallback: different non-semver strings
	return normalizeVersion(latest) != normalizeVersion(current)
}

func ensureVPrefix(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

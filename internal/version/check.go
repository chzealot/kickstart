package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
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

	hasUpdate := latest != "" && normalizeVersion(latest) != normalizeVersion(Version)
	c.result <- &CheckResult{
		Latest:    latest,
		Current:   Version,
		HasUpdate: hasUpdate,
	}
}

func fetchLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/chzealot/kickstart/releases/latest")
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

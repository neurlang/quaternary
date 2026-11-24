package v1

import (
	"fmt"
	"os"
	"testing"
)

// TestMain fails the run if any test, fuzz target or benchmark leaves the global
// Integration hook set. The hook changes how every key in every filter is
// hashed, so a leaked hook silently changes the result of whatever runs next in
// the same process; tests must be hermetic. Go happens to run fuzz seed corpora
// after all Test functions, which used to hide such a leak.
func TestMain(m *testing.M) {
	code := m.Run()
	if Integration != nil {
		fmt.Fprintln(os.Stderr, "FAIL: a test left the package-level Integration hook set; restore it with setIntegrationForTest")
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

// setIntegrationForTest installs h as the Integration hook and restores the
// previous value when tb (a *testing.T, *testing.B or *testing.F) finishes.
func setIntegrationForTest(tb testing.TB, h Hashstructure) {
	tb.Helper()
	prev := Integration
	Integration = h
	tb.Cleanup(func() { Integration = prev })
}

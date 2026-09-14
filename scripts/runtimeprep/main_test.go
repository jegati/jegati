package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for name, n := range map[string]int{"access": 3, "tunnel": 1} {
		dir := filepath.Join(root, "cmd/cloudflared", name)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cmd.go"), []byte(strings.Repeat("sentry.Init(sentry.ClientOptions{})\n", n)), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
func TestPatchDisablesAllReviewedInitializers(t *testing.T) {
	root := fixture(t)
	if err := patch(root); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"access", "tunnel"} {
		raw, err := os.ReadFile(filepath.Join(root, "cmd/cloudflared", name, "cmd.go"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "sentry.Init(") || !strings.Contains(string(raw), "func gatiSentryDisabled(_ sentry.ClientOptions) error { return nil }") {
			t.Fatal("reporting initializer remains")
		}
	}
	if patch(root) == nil {
		t.Fatal("double application must fail")
	}
}
func TestPatchRejectsUnexpectedInitializerBeforeMutation(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, "extra.go"), []byte("sentry.Init(options)"), 0o600); err != nil {
		t.Fatal(err)
	}
	if patch(root) == nil {
		t.Fatal("new initializer was silently permitted")
	}
	raw, err := os.ReadFile(filepath.Join(root, "cmd/cloudflared/access/cmd.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "sentry.Init(") {
		t.Fatal("source mutated before review")
	}
}

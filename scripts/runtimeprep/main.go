// Apply the narrowly scoped cloudflared privacy patch to a verified source export.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func patch(root string) error {
	counts := map[string]int{"cmd/cloudflared/access/cmd.go": 3, "cmd/cloudflared/tunnel/cmd.go": 1}
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(raw), "sentry.Init(") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if strings.Count(string(raw), "sentry.Init(") != counts[filepath.ToSlash(rel)] {
			return fmt.Errorf("unreviewed Sentry initializer: %s", rel)
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return err
	}
	if len(paths) != len(counts) {
		return fmt.Errorf("upstream Sentry initialization changed")
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result := strings.ReplaceAll(string(raw), "sentry.Init(", "gatiSentryDisabled(")
		result += "\n// GATI: never initialize the third-party crash-reporting client.\nfunc gatiSentryDisabled(_ sentry.ClientOptions) error { return nil }\n"
		if err := os.WriteFile(path, []byte(result), 0o644); err != nil {
			return err
		}
	}
	return nil
}
func main() {
	if len(os.Args) == 3 && os.Args[1] == "licenses" {
		if err := licenses(os.Args[2]); err != nil {
			panic(err)
		}
		return
	}
	if len(os.Args) != 2 {
		panic("expected verified source directory")
	}
	if err := patch(os.Args[1]); err != nil {
		panic(err)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Retain dependency license/notice files and a module inventory in shipped images.
func licenses(out string) error {
	raw, err := exec.Command("go", "list", "-m", "-json", "all").Output()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(out, "modules.jsonl"), raw, 0o644); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	type module struct {
		Path, Version, Dir string
		Main               bool
		Replace            *struct{ Path, Version, Dir string }
	}
	for {
		var m module
		if err = decoder.Decode(&m); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if m.Main {
			continue
		}
		if m.Replace != nil {
			m.Path, m.Version, m.Dir = m.Replace.Path, m.Replace.Version, m.Replace.Dir
		}
		if m.Dir == "" {
			continue
		} // go mod download all populates directories before this command.
		err = filepath.WalkDir(m.Dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			name := strings.ToUpper(entry.Name())
			if !strings.HasPrefix(name, "LICENSE") && !strings.HasPrefix(name, "COPYING") && !strings.HasPrefix(name, "NOTICE") {
				return nil
			}
			rel, err := filepath.Rel(m.Dir, path)
			if err != nil {
				return err
			}
			target := filepath.Join(out, m.Path+"@"+m.Version, rel)
			if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, body, 0o644)
		})
		if err != nil {
			return err
		}
	}
}

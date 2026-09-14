package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrivateDirectoryAndExistingKeyRetained(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "private")
	if err := run(directory); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "vapid.json")
	first, err := os.ReadFile(path)
	if err != nil || len(first) == 0 {
		t.Fatal("no generated service key")
	}
	info, err := os.Stat(directory)
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatal("key directory is not private")
	}
	if err := run(directory); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil || string(first) != string(second) {
		t.Fatal("existing key changed")
	}
}

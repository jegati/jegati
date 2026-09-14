// gati-push-keys creates project-local service keys without printing secrets.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jegati/jegati/internal/notification"
	"os"
	"path/filepath"
)

func main() {
	directory := flag.String("directory", ".runtime", "private operational key directory")
	flag.Parse()
	if e := run(*directory); e != nil {
		fmt.Fprintln(os.Stderr, "Push key setup failed; existing files were not replaced.")
		os.Exit(1)
	}
}
func run(directory string) error {
	if e := os.MkdirAll(directory, 0700); e != nil {
		return e
	}
	if e := os.Chmod(directory, 0700); e != nil {
		return e
	}
	keys, e := notification.GenerateKeys()
	if e != nil {
		return e
	}
	file, e := os.OpenFile(filepath.Join(directory, "vapid.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if os.IsExist(e) {
		fmt.Println("Existing project-local push key file retained.")
		return nil
	}
	if e != nil {
		return e
	}
	if e = json.NewEncoder(file).Encode(keys); e != nil {
		file.Close()
		os.Remove(filepath.Join(directory, "vapid.json"))
		return e
	}
	if e = file.Close(); e != nil {
		return e
	}
	fmt.Println("Project-local push service keys created. Secret values are not printed.")
	return nil
}

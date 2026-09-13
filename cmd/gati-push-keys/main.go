// gati-push-keys creates project-local service keys without printing secrets.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/jegati/jegati/internal/notification"
	"os"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "Push key setup failed; existing files were not replaced.")
		os.Exit(1)
	}
}
func run() error {
	if e := os.MkdirAll(".runtime", 0700); e != nil {
		return e
	}
	if e := os.Chmod(".runtime", 0700); e != nil {
		return e
	}
	keys, e := notification.GenerateKeys()
	if e != nil {
		return e
	}
	file, e := os.OpenFile(".runtime/vapid.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if os.IsExist(e) {
		fmt.Println("Existing project-local push key file retained.")
		return nil
	}
	if e != nil {
		return e
	}
	if e = json.NewEncoder(file).Encode(keys); e != nil {
		file.Close()
		os.Remove(".runtime/vapid.json")
		return e
	}
	if e = file.Close(); e != nil {
		return e
	}
	fmt.Println("Project-local push service keys created. Secret values are not printed.")
	return nil
}

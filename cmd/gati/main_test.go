package main

import "testing"

func TestSimulationCannotBindExternally(t *testing.T) {
	for _, address := range []string{"0.0.0.0:8080", ":8080", "localhost:8080", "192.168.1.1:8080", "[::]:8080"} {
		if validateBind(address, true) == nil {
			t.Fatalf("accepted simulation bind %s", address)
		}
	}
	for _, address := range []string{"127.0.0.1:8080", "[::1]:8080"} {
		if err := validateBind(address, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateBind("0.0.0.0:8080", false); err != nil {
		t.Fatal(err)
	}
}

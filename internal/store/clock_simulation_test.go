//go:build simulation

package store

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSimulationClockAndNamespace(t *testing.T) {
	if os.Getenv("GATI_SIM_INTEGRATION") != "1" {
		t.Skip("requires isolated simulation store")
	}
	raw, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	s, e := Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(raw)))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Client.Close()
	if e = s.Client.Set(ctx, "gati:s:forbidden", "value", time.Minute).Err(); e == nil {
		t.Fatal("simulation credential can write production namespace")
	}
	hash := Hash([]byte("synthetic-clock-case"))
	signal, e := s.Create(ctx, hash, "tirana-v1:1000:5:5", 3, 30, time.Minute, 1000)
	if e != nil {
		t.Fatal(e)
	}
	time.Sleep(10 * time.Millisecond)
	again, e := s.Create(ctx, hash, signal.Cell, 3, 30, time.Minute, 1000)
	if e != nil || again.ExpiresAt != signal.ExpiresAt {
		t.Fatal("retry changed simulated deadline")
	}
	if _, e = s.AdvanceClock(ctx, -1); e == nil {
		t.Fatal("clock moved backwards")
	}
	if _, e = s.AdvanceClock(ctx, 60_000); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Status(ctx, hash); e != ErrGone {
		t.Fatalf("expired read = %v", e)
	}
	if _, e = s.Create(ctx, hash, signal.Cell, 3, 30, time.Minute, 1000); e != ErrGone {
		t.Fatalf("expired retry = %v", e)
	}
	if _, e = s.Cleanup(ctx, 1000); e != nil {
		t.Fatal(e)
	}
	count, e := s.Client.ZCard(ctx, keyPrefix+"cell:"+signal.Cell).Result()
	if e != nil || count != 0 {
		t.Fatal("expired index members remain")
	}
}

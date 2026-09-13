package store

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func connectTest(t testing.TB) *Store {
	t.Helper()
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("real Valkey test: use make test-store")
	}
	password, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal("test credential file unavailable")
	}
	var s *Store
	for attempt := 0; attempt < 20; attempt++ {
		s, e = Connect(context.Background(), os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(password)))
		if e == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Client.Close() })
	return s
}
func fresh() string { b := make([]byte, 32); rand.Read(b); return Hash(b) }
func TestRealStoreExpiryCancellationAndIdempotency(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	hash := fresh()
	cell := "tirana-v1:1000:4:4"
	first, e := s.Create(ctx, hash, cell, 3, 30, 2*time.Second, 1000)
	if e != nil {
		t.Fatal(e)
	}
	retry, e := s.Create(ctx, hash, cell, 3, 30, 2*time.Second, 1000)
	if e != nil || retry.ExpiresAt != first.ExpiresAt {
		t.Fatal("retry extended session")
	}
	if _, e = s.Create(ctx, hash, cell, 5, 30, time.Second, 1000); !errors.Is(e, ErrConflict) {
		t.Fatal("reused capability accepted changed signal")
	}
	if e = s.Cancel(ctx, hash); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Status(ctx, hash); !errors.Is(e, ErrGone) {
		t.Fatal("cancelled signal still readable")
	}
	if _, e = s.Create(ctx, hash, cell, 3, 30, time.Second, 1000); !errors.Is(e, ErrGone) {
		t.Fatal("cancelled credential replay resurrected signal")
	}
	if e = s.Cancel(ctx, hash); e != nil {
		t.Fatal(e)
	}
	if n, _ := s.Client.ZCard(ctx, "gati:cell:"+cell).Result(); n != 0 {
		t.Fatal("cancel left live index membership")
	}
	survivor := fresh()
	s.Create(ctx, survivor, cell, 3, 30, 3*time.Second, 1000)
	defer s.Cancel(ctx, survivor)
	expiring := fresh()
	if _, e = s.Create(ctx, expiring, cell, 3, 30, 100*time.Millisecond, 1000); e != nil {
		t.Fatal(e)
	}
	time.Sleep(150 * time.Millisecond)
	if _, e = s.Status(ctx, expiring); !errors.Is(e, ErrGone) {
		t.Fatal("expired signal readable")
	}
	if _, e = s.Cleanup(ctx, 100); e != nil {
		t.Fatal(e)
	}
	if n, _ := s.Client.ZCard(ctx, "gati:cell:"+cell).Result(); n != 1 {
		t.Fatal("expired index member retained or survivor lost")
	}
	if ttl, _ := s.Client.PTTL(ctx, "gati:cell:"+cell).Result(); ttl <= 0 || ttl > 3*time.Second {
		t.Fatal("index has no bounded TTL")
	}
}
func TestRealStoreConcurrentCreationAndACL(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	hash := fresh()
	cell := "tirana-v1:1000:3:3"
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := s.Create(ctx, hash, cell, 3, 30, time.Second, 1000); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	if n, _ := s.Client.ZCard(ctx, "gati:cell:"+cell).Result(); n != 1 {
		t.Fatal("concurrent retries inflated index")
	}
	defer s.Cancel(ctx, hash)
	for _, cmd := range [][]any{{"CONFIG", "GET", "*"}, {"SAVE"}, {"BGSAVE"}, {"MONITOR"}, {"ACL", "LIST"}, {"SET", "outside-gati", "x"}, {"GET", "outside-gati"}} {
		if e := s.Client.Do(ctx, cmd...).Err(); e == nil || !strings.Contains(e.Error(), "NOPERM") {
			t.Fatalf("restricted command not denied: %v", cmd[0])
		}
	}
	key, e := s.NetworkKey(ctx, "synthetic-network")
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 4; i++ {
		allowed, e := s.Allow(ctx, "test:"+key, 3, time.Second)
		if e != nil || allowed != (i < 3) {
			t.Fatal("limiter failed")
		}
	}
}

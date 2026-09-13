package store

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRealNetworkSecretEpochBoundary(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	prefix := keyPrefix + "rate:secret:test:" + fresh() + ":"
	defer s.Client.Del(ctx, prefix+"9", prefix+"10")
	first, next := strings.Repeat("a", 32), strings.Repeat("b", 32)
	// Substitute clock values in this test's script only: one millisecond before
	// rotation, then the next epoch. Production accepts no clock argument.
	before := redis.NewScript(strings.Replace(networkSecretLua, "redis.call('TIME')", "{5999,999000}", 1))
	result, err := before.Run(ctx, s.Client, []string{prefix}, first).StringSlice()
	if err != nil || len(result) != 2 || result[0] != "9" || result[1] != first {
		t.Fatal("atomic read failed immediately before secret expiration")
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := s.Client.Get(ctx, prefix+"9").Result(); err != redis.Nil {
		t.Fatal("previous epoch secret survived its deadline")
	}
	after := redis.NewScript(strings.Replace(networkSecretLua, "redis.call('TIME')", "{6000,0}", 1))
	result, err = after.Run(ctx, s.Client, []string{prefix}, next).StringSlice()
	if err != nil || len(result) != 2 || result[0] != "10" || result[1] != next {
		t.Fatal("new epoch did not return its own secret")
	}
	// Other API replicas proposing different randomness must share the winner.
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := after.Run(ctx, s.Client, []string{prefix}, first).StringSlice()
			if e != nil || len(v) != 2 || v[1] != next {
				t.Error("concurrent replicas disagreed on the epoch secret")
			}
		}()
	}
	wg.Wait()
	if ttl, err := s.Client.PTTL(ctx, prefix+"10").Result(); err != nil || ttl <= 0 || ttl > 10*time.Minute {
		t.Fatal("rotation secret has no bounded lifetime")
	}
}

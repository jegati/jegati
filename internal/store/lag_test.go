package store

import (
	"context"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestDeadlineLagWithExpiredIndexAndNoRecord(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	now, err := s.Now(ctx)
	if err != nil {
		t.Fatal(err)
	}
	key := keyPrefix + "pending"
	member := fresh()
	if err := s.Client.ZAdd(ctx, key, redis.Z{Score: float64(now - 7000), Member: member}).Err(); err != nil {
		t.Fatal(err)
	}
	s.Client.PExpire(ctx, key, time.Minute)
	t.Cleanup(func() { s.Client.ZRem(ctx, key, member) })
	lag, err := s.PendingLag(ctx)
	if err != nil || lag < 7*time.Second {
		t.Fatal("overdue index hidden", lag, err)
	}
}

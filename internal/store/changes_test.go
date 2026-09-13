package store

import (
	"context"
	"testing"
	"time"
)

func TestCellChangesDoNotLoseLaterWritesAndLeaseOwnership(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	cell := "tirana-v1:100:51:51"
	hash := fresh()
	defer s.Cancel(ctx, hash)
	s.TakeChangedCells(ctx, 20000)
	if _, err := s.Create(ctx, hash, cell, 1, 30, time.Minute, 10000); err != nil {
		t.Fatal(err)
	}
	cells, err := s.TakeChangedCells(ctx, 20000)
	if err != nil || len(cells) != 1 || cells[0] != cell {
		t.Fatal("creation not scheduled", err)
	}
	if err := s.Cancel(ctx, hash); err != nil {
		t.Fatal(err)
	}
	cells, err = s.TakeChangedCells(ctx, 20000)
	if err != nil || len(cells) != 1 || cells[0] != cell {
		t.Fatal("change during previous batch lost", err)
	}
	if ttl := s.Client.PTTL(ctx, keyPrefix+"dirty").Val(); ttl <= 0 || ttl > 10*time.Second {
		t.Fatal("dirty marker lifecycle")
	}
	owner := fresh()
	s.Client.Del(ctx, keyPrefix+"worker:lease")
	won, newOwner, err := s.MatchingLeaseState(ctx, owner, 1000)
	if err != nil || !won || !newOwner {
		t.Fatal("new lease not distinguished")
	}
	won, newOwner, err = s.MatchingLeaseState(ctx, owner, 1000)
	if err != nil || !won || newOwner {
		t.Fatal("renewal not distinguished")
	}
	won, _, err = s.MatchingLeaseState(ctx, fresh(), 1000)
	if err != nil || won {
		t.Fatal("another writer admitted")
	}
	s.Client.Del(ctx, keyPrefix+"worker:lease")
}

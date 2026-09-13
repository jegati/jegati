package store

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPushLifecycleClaimsAndCancellation(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	p, grid := proposalTest(t, s)
	hash := p.Founders[0].Hash
	before, e := s.Status(ctx, hash)
	if e != nil {
		t.Fatal(e)
	}
	binding := PushBinding{Binding: fresh(), Digest: fresh(), Ciphertext: "synthetic-encrypted-transport", ExpiresAt: before.ExpiresAt + 60000}
	expiry, e := s.RegisterPush(ctx, hash, binding)
	if e != nil || expiry != before.ExpiresAt {
		t.Fatal("subscription exceeded willingness", e)
	}
	if again, e := s.RegisterPush(ctx, hash, binding); e != nil || again != expiry {
		t.Fatal("registration retry changed lifetime")
	}
	other := binding
	other.Binding = fresh()
	if _, e := s.RegisterPush(ctx, hash, other); e != ErrConflict {
		t.Fatal("replacement lacked explicit opt-out")
	}
	policy := PushPolicy{PollMS: 10, QueueMS: 5000, GapMS: 5000, RetryMS: 10, LeaseMS: 1000, MinimumMS: 1000, MaxAttempts: 3}
	if job, e := s.ClaimPush(ctx, hash, fresh(), policy); e != nil || job != nil {
		t.Fatal("notification created before invitation", e)
	}
	after, _ := s.Status(ctx, hash)
	if after.State != before.State || after.Gathering != "" || after.ArrivalUntil != 0 {
		t.Fatal("subscription changed participation")
	}
	id := fresh()
	if _, e = s.Reserve(ctx, id, "test", p, grid, 20, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	time.Sleep(30 * time.Millisecond)
	if _, e = s.Activate(ctx, id, "test", 1000, 60000); e != nil {
		t.Fatal(e)
	}
	results := make(chan *PushJob, 20)
	failures := make(chan error, 20)
	var wg sync.WaitGroup
	for n := 0; n < 20; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); j, e := s.ClaimPush(ctx, hash, fresh(), policy); results <- j; failures <- e }()
	}
	wg.Wait()
	close(results)
	close(failures)
	for e := range failures {
		if e != nil {
			t.Fatal(e)
		}
	}
	var claimed *PushJob
	count := 0
	for j := range results {
		if j != nil {
			count++
			claimed = j
		}
	}
	if count != 1 || claimed.Until > expiry || claimed.Ciphertext != binding.Ciphertext {
		t.Fatal("nonexclusive or unbounded delivery claim")
	}
	if e = s.FinishPush(ctx, hash, claimed.Claim, "retry", 10); e != nil {
		t.Fatal(e)
	}
	time.Sleep(20 * time.Millisecond)
	retry, e := s.ClaimPush(ctx, hash, fresh(), policy)
	if e != nil || retry == nil || retry.Claim == claimed.Claim {
		t.Fatal("bounded retry missing", e)
	}
	if e = s.FinishPush(ctx, hash, claimed.Claim, "gone", 10); e != nil {
		t.Fatal(e)
	}
	if s.Client.Exists(ctx, keyPrefix+"push:"+hash).Val() != 1 {
		t.Fatal("stale acknowledgement deleted a newer claim")
	}
	if e = s.FinishPush(ctx, hash, retry.Claim, "sent", 10); e != nil {
		t.Fatal(e)
	}
	if job, e := s.ClaimPush(ctx, hash, fresh(), policy); e != nil || job != nil {
		t.Fatal("same event delivered twice after ack")
	}
	if e = s.DropPush(ctx, hash); e != nil {
		t.Fatal(e)
	}
	if _, e = s.RegisterPush(ctx, hash, other); e != nil {
		t.Fatal(e)
	}
	if job, e := s.ClaimPush(ctx, hash, fresh(), policy); e != nil || job != nil {
		t.Fatal("opt-out/opt-in bypassed notification gap")
	}
	// The other founder exercises the attempt ceiling without provider traffic.
	second := p.Founders[1].Hash
	other.ExpiresAt = expiry
	if _, e = s.RegisterPush(ctx, second, other); e != nil {
		t.Fatal(e)
	}
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		job, e := s.ClaimPush(ctx, second, fresh(), policy)
		if e != nil || job == nil {
			t.Fatal("retry attempt missing", e)
		}
		if e = s.FinishPush(ctx, second, job.Claim, "retry", 1); e != nil {
			t.Fatal(e)
		}
		time.Sleep(time.Duration(policy.RetryMS*(1<<attempt)+10) * time.Millisecond)
	}
	if job, e := s.ClaimPush(ctx, second, fresh(), policy); e != nil || job != nil {
		t.Fatal("delivery exceeded retry ceiling")
	}
	if e = s.Cancel(ctx, hash); e != nil {
		t.Fatal(e)
	}
	if s.Client.Exists(ctx, keyPrefix+"push:"+hash, keyPrefix+"push-gap:"+hash).Val() != 0 {
		t.Fatal("cancel retained subscription/quota")
	}
	if job, e := s.ClaimPush(ctx, hash, fresh(), policy); e != nil || job != nil {
		t.Fatal("cancelled participant still queued")
	}
	if _, e = s.RegisterPush(ctx, hash, binding); e != ErrGone {
		t.Fatal("cancelled subscription resurrected")
	}
}

func TestPushNativeExpiry(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	hash := fresh()
	signal, e := s.Create(ctx, hash, "tirana-v1:1000:5:5", 3, 30, 60*time.Millisecond, 1000)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.RegisterPush(ctx, hash, PushBinding{Binding: fresh(), Digest: fresh(), Ciphertext: "synthetic", ExpiresAt: signal.ExpiresAt})
	if e != nil {
		t.Fatal(e)
	}
	time.Sleep(80 * time.Millisecond)
	due, e := s.PushDue(ctx, 1000)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, v := range due {
		if v == hash {
			found = true
		}
	}
	// The whole index may already have expired; either case must fail closed.
	if found {
		if job, e := s.ClaimPush(ctx, hash, fresh(), PushPolicy{}); e != nil || job != nil {
			t.Fatal("expired subscription delivered")
		}
	}
	if s.Client.Exists(ctx, keyPrefix+"push:"+hash).Val() != 0 {
		t.Fatal("native subscription TTL missing")
	}
}

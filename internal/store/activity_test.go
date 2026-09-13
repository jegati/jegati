package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/config"
)

func TestRealActivityCaptureCancellationAndBoundedReads(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	hashes := []string{}
	cell := "tirana-v1:100:12:12"
	// Force many pages, including repeated creates that must count only once.
	c.Limits.CleanupBatchSize = 7
	for i := 0; i < 55; i++ {
		hash := fresh()
		hashes = append(hashes, hash)
		if _, e = s.Create(ctx, hash, cell, 3, 30, time.Minute, c.Limits.MaxActiveSignals); e != nil {
			t.Fatal(e)
		}
	}
	t.Cleanup(func() {
		for _, h := range hashes {
			s.Cancel(ctx, h)
		}
	})
	s.Create(ctx, hashes[0], cell, 3, 30, time.Minute, c.Limits.MaxActiveSignals)
	r, data, e := s.CaptureActivity(ctx, c)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, a := range r.Areas {
		if a.Cell == "tirana-v1:1000:1:1" {
			found = a.Willing == 50
		}
	}
	if !found {
		t.Fatal("paging lost or duplicated willingness")
	}
	for _, h := range hashes {
		if strings.Contains(string(data), h) {
			t.Fatal("credential hash in public output")
		}
	}
	for _, h := range hashes[:36] {
		if e = s.Cancel(ctx, h); e != nil {
			t.Fatal(e)
		}
	}
	r, _, e = s.CaptureActivity(ctx, c)
	if e != nil {
		t.Fatal(e)
	}
	for _, a := range r.Areas {
		if a.Cell == "tirana-v1:1000:1:1" {
			t.Fatal("cancelled population still published")
		}
	}
	// Churn is an observation interval: no duplicate inputs or unbounded scan.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, h := range hashes[36:] {
			s.Cancel(ctx, h)
		}
	}()
	_, _, e = s.CaptureActivity(ctx, c)
	wg.Wait()
	if e != nil {
		t.Fatal(e)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, e = s.CaptureActivity(canceled, c); e == nil {
		t.Fatal("cancelled capture published")
	}
}
func TestRealActivityFencingReleaseImmutabilityAndExpiry(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	// A distinct nonsecret config hash isolates this test's aggregate keys.
	c.Limits.MaxActiveSignals = 123456
	now, e := s.Now(ctx)
	if e != nil {
		t.Fatal(e)
	}
	_, hash := c.Canonical()
	interval := int64(c.PublicActivity.ReleaseSeconds) * 1000
	epoch := now/interval - int64(c.PublicActivity.DelayEpochs) - 1
	r := activity.Release{Version: 1, ID: fmt.Sprint(epoch), ConfigHash: hash, ObservedFrom: epoch * interval, ObservedUntil: epoch * interval, ReleaseAt: now + 150, ExpiresAt: now + 500, Areas: []activity.Area{}, Gatherings: []activity.Event{}}
	data, _ := json.Marshal(r)
	owner := fresh()
	s.Client.Del(ctx, keyPrefix+"activity:lease", activityKey(hash, epoch))
	defer s.Client.Del(ctx, keyPrefix+"activity:lease", activityKey(hash, epoch))
	if ok, e := s.ActivityLease(ctx, owner, 1000); e != nil || !ok {
		t.Fatal("lease unavailable", e)
	}
	if ok, e := s.StageActivity(ctx, "wrong-owner", c, r, data); e != nil || ok {
		t.Fatal("unfenced publisher accepted", e)
	}
	if ok, e := s.StageActivity(ctx, owner, c, r, data); e != nil || !ok {
		t.Fatal("stage failed", e)
	}
	if bytes, _, e := s.LatestActivity(ctx, c); e != nil || len(bytes) != 0 {
		t.Fatal("early release", e)
	}
	changed := r
	changed.ID = "amended"
	changedData, _ := json.Marshal(changed)
	if ok, e := s.StageActivity(ctx, owner, c, changed, changedData); e != nil || ok {
		t.Fatal("epoch overwritten", e)
	}
	time.Sleep(180 * time.Millisecond)
	bytes, _, e := s.LatestActivity(ctx, c)
	if e != nil || string(bytes) != string(data) {
		t.Fatal("release/restart read failed", e)
	}
	time.Sleep(350 * time.Millisecond)
	if bytes, _, e = s.LatestActivity(ctx, c); e != nil || len(bytes) != 0 {
		t.Fatal("expired release visible", e)
	}
	if n, _ := s.Client.Exists(ctx, activityKey(hash, epoch)).Result(); n != 0 {
		t.Fatal("aggregate bytes outlive TTL")
	}
}

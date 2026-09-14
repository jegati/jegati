package worker

import (
	"testing"
	"time"
)

func TestFastMissingPublicationAccumulatesAlertableLag(t *testing.T) {
	for _, v := range []struct {
		now int64
		lag time.Duration
	}{
		{605000, 0}, {610000, 10 * time.Second}, {620000, 20 * time.Second}, {630000, 30 * time.Second},
	} {
		if got := publicationLag(v.now, 60, 10000, 0, nil); got != v.lag {
			t.Fatalf("at %d lag %s want %s", v.now, got, v.lag)
		}
	}
	if got := publicationLag(630000, 60, 10000, 0, []byte(`{"observed_from":620000}`)); got != 0 {
		t.Fatal("fresh release reported behind")
	}
	if got := publicationLag(650000, 60, 10000, 0, []byte(`{"observed_from":600000}`)); got != 40*time.Second {
		t.Fatal("stale release lag lost")
	}
}

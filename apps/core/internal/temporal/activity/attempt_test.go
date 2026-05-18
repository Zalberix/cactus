package activity

import "testing"

func TestActivityAttemptNumberUsesTemporalAttemptAsIs(t *testing.T) {
	if got := activityAttemptNumber(1); got != 1 {
		t.Fatalf("first Temporal attempt must stay 1, got %d", got)
	}
}

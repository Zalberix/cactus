package activity

import "testing"

func TestActivityAttemptNumberUsesTemporalAttemptAsIs(t *testing.T) {
	if got := activityAttemptNumber(0); got != 1 {
		t.Fatalf("first Temporal attempt must be converted to 1, got %d", got)
	}
}

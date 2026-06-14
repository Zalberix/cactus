package gum

import "testing"

func TestFormatCommandQuotesArguments(t *testing.T) {
	got := FormatCommand("style", "--foreground", "212", "hello world")
	want := `go tool gum "style" "--foreground" "212" "hello world"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

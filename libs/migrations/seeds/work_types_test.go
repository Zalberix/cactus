package seeds

import "testing"

func TestHTMLBootstrapTokenHashConstant(t *testing.T) {
	got := htmlBootstrapTokenHash
	want := "fcaf410de830d072b301a2e7ca1092320647584c2585c9b5e5f64cebf1990164"
	if got != want {
		t.Fatalf("htmlBootstrapTokenHash = %q, want %q", got, want)
	}
}

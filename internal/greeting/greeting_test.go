package greeting

import (
	"os"
	"testing"
)

func TestForNamePreservesUnicode(t *testing.T) {
	want, err := os.ReadFile("testdata/unicode.golden")
	if err != nil {
		t.Fatal(err)
	}
	if got := ForName("世界") + "\n"; got != string(want) {
		t.Errorf("ForName(name) plus newline = %q; want %q", got, want)
	}
}

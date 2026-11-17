package gopkgtemp_test

import (
	"testing"

	gopkgtemp "github.com/oto-lab/go-pkg-temp"
)

func TestGreet(t *testing.T) {
	for _, tt := range []struct{ name, input, want string }{
		{name: "named", input: "Go", want: "Hello, Go!"},
		{name: "default", want: "Hello, world!"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := gopkgtemp.Greet(tt.input); got != tt.want {
				t.Errorf("Greet(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

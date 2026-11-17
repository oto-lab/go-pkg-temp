package gopkgtemp

import "github.com/oto-lab/go-pkg-temp/internal/greeting"

// Greet returns a greeting addressed to name, or to "world" when name is empty.
// For example, Greet("Go") returns "Hello, Go!".
func Greet(name string) string {
	return greeting.ForName(name)
}

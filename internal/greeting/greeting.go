// Package greeting formats the library's sample greeting.
package greeting

import "cmp"

// ForName returns "Hello, <name>!", using "world" for an empty name.
func ForName(name string) string {
	return "Hello, " + cmp.Or(name, "world") + "!"
}

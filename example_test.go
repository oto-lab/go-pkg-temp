package gopkgtemp_test

import (
	"fmt"

	gopkgtemp "github.com/oto-lab/go-pkg-temp"
)

func ExampleGreet() {
	fmt.Println(gopkgtemp.Greet("Go"))
	// Output: Hello, Go!
}

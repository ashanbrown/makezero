package always

import "fmt"

func Foo() {
	x := make([]int, 5) // want "slice `x` does not have non-zero initial length"
	fmt.Println(x)
}

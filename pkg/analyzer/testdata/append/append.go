package append

import "fmt"

func Foo() {
	x := make([]int, 5)
	x = append(x, 1) // want "append to slice `x` with non-zero initialized length"
	fmt.Println(x)
}

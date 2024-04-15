package appendwithcopy

import "fmt"

func Foo() {
	in := []int{1, 2, 3}
	out := make([]int, len(in))
	// first copy and then append
	copy(out, in)
	out = append(out, 4)
	fmt.Println(out)
}

func Foo2() {
	in := []int{1, 2, 3}
	out := make([]int, len(in))
	out = append(out, 4) // want "append to slice `out` with non-zero initialized length"
	copy(out, in)
	fmt.Println(out)
}

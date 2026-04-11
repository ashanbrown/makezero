package append

import (
	"fmt"
	"typespkg"
)

func Foo() {
	x := make([]int, 5)
	x = append(x, 1) // want "append to slice `x` with non-zero initialized length"
	fmt.Println(x)
}

func Bar() {
	x := make(typespkg.S, 10)
	x = append(x, "...") // want "append to slice `x` with non-zero initialized length"
	fmt.Println(x)
}

func Baz() {
	x := make(typespkg.M, 10)
	fmt.Println(x)
}

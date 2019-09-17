package examples

import "fmt"

func Foo() {
	x := make([]int, 5)
	x = append(x, 1)
	fmt.Println(x)
}

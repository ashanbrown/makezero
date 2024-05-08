package appendwithcopy

import (
	"fmt"
	"strconv"
)

func Foo() {
	in := []int{1, 2, 3}
	out := make([]int, len(in))
	copy(out, in)
	out = append(out, 4) // want "append to slice `out` with non-zero initialized length, and called by funcs: copy"
	fmt.Println(out)
}

func Foo2() {
	in := []int{1, 2, 3}
	out := make([]int, len(in))
	out = append(out, 4) // want "append to slice `out` with non-zero initialized length"
	copy(out, in)
	fmt.Println(out)
}

func Foo3() {
	in := []byte{1, 2, 3}
	out := make([]byte, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = in[i] + 1
	}
	out = append(out, 4) // want "append to slice `out` with non-zero initialized length, and has assigned by index"
	copy(out, in)
	out = append(out, 5) // want "append to slice `out` with non-zero initialized length, and called by funcs: copy, and has assigned by index"

	out = Decode(1000, out)
	out = append(out, 0) // want "append to slice `out` with non-zero initialized length, and called by funcs: copy,Decode, and has assigned by index"
	fmt.Println(out)
}

func Decode(num int64, out []byte) []byte {
	return strconv.AppendInt(out, num, 10)
}

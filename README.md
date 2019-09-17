# makezero

makezero is a Go static analysis tool to find slice declarations that are not initialized with zero length and are later
used with append.

## Installation

    go get -u github.com/ashanbrown/makezero

## Usage

Similar to other Go static analysis tools (such as golint, go vet), makezero can be invoked with one or more filenames, directories, or packages named by its import path. makezero also supports the `...` wildcard.

    makezero [-always] packages...

### Flags
- **-set_exit_status** (default false) - Set exit status to 1 if any issues are found.

## Purpose

To prevent bugs caused by initializing a slice with non-constant length and later appending to it.  The recommended
[prealloc](https://github.com/alexkohler/prealloc) linter wisely encourages the developer to pre-allocate, but when we preallocate a slice with empty values in it and later append to it, we can easily introduce extra empty element in that slice.

Consider the case below:

```Go
import "testing"

func copyNumbers(nums []int) []int {
  values := make([]int, len(num)) // satisfy prealloc
  for _, n := range nums {
    values = apppend(values, n)
  }
  return values 
}

```

In this case, you probably mean to preallocate with length 0 `values := make([]int, 0, len(num))`.

## TODO

Consider whether this should be part of prealloc itself.

## Contributing

Pull requests welcome!

# makezero

[![CircleCI](https://circleci.com/gh/ashanbrown/makezero/tree/master.svg?style=svg)](https://circleci.com/gh/ashanbrown/makezero/tree/master)

makezero is a Go static analysis tool to find slice declarations that are not initialized with zero length and are later
used with append.

## Installation

    go get -u github.com/ashanbrown/makezero

## Usage

Similar to other Go static analysis tools (such as golint, go vet), makezero can be invoked with one or more filenames, directories, or packages named by its import path. makezero also supports the `...` wildcard.

    makezero [-always] packages...

### Flags
- **-set_exit_status** (default false) - Set exit status to 1 if any issues are found.
- **-always** (default false) - Always require slices to be initialized with zero length, regardless of whether they are used with append.

## Purpose

To prevent bugs caused by initializing a slice with non-constant length and later appending to it.  The recommended
[prealloc](https://github.com/alexkohler/prealloc) linter wisely encourages the developer to pre-allocate, but when we preallocate a slice with empty values in it and later append to it, we can easily introduce extra empty element in that slice.

Consider the case below:

```Go
func copyNumbers(nums []int) []int {
  values := make([]int, len(nums)) // satisfy prealloc
  for _, n := range nums {
    values = append(values, n)
  }
  return values
}
```

In this case, you probably mean to preallocate with length 0 `values := make([]int, 0, len(nums))`.

The `-always` directive enforces that slice created with `make` always have initial length of zero.  This may sound
draconian but it encourages the use of `append` when building up arrays rather than C-style code featuring the index
variable `i` such as in:

```Go
func copyNumbers(nums []int) []int {
  values := make([]int, len(nums))
  for i, n := range nums {
    values[i] = n
  }
  return values
}

```

## Ignoring issues

You can ignore a particular issue by including the directive `// nozero` on that line

## TODO

Consider whether this should be part of prealloc itself.

## Contributing

Pull requests welcome!

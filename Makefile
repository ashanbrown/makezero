SHELL=bash

test:
	diff <(sed 's|CURDIR|$(CURDIR)|' examples/expected_results.txt) <(go run . -always ./examples 2>&1)

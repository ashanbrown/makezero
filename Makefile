SHELL=bash

test:
	diff examples/expected_results.txt <(go run . -always ./examples 2>&1)

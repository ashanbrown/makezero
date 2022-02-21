SHELL=bash

test:
	diff <(sed 's|CURDIR|$(CURDIR)|' examples/expected_results.txt) <(go run . -always ./examples 2>&1)
	diff <(sed 's|CURDIR|$(CURDIR)|' cmd/makezero/examples/expected_results.txt) <(go run ./cmd/makezero -always ./examples 2>&1)

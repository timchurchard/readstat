.PHONY:

generate:
	./.github/mockgen-version.sh
	go generate -v ./...

lint: generate
	./.github/lint-version.sh
	golangci-lint run -v ./...

test: generate
	go test -cover ./...


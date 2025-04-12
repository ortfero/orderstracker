project := "orderstracker"

default: test

fmt:
    go fmt ./...

vet:
    go vet ./...

test:
    go test -v ./...

bench:
    go test -bench=.

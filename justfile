default:
    @just --list

build:
    go build ./...

test:
    go test ./...

vet:
    go vet ./...

fmt:
    go fmt ./...

example:
    go run ./example

clean:
    rm -f /tmp/bezel-example

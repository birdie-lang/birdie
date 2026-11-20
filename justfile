set default-list := true

# Start the Birdie REPL.
@repl:
    go run ./cmd/birdie

# Run all Go tests.
@test:
    go test ./...

# Check Go code for common mistakes.
@vet:
    go vet ./...

# Format Go code.
@fmt:
    go fmt ./...

# Run vet and tests.
@check: vet test

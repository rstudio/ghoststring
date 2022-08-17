default: vet test

vet:
  go vet ./...

test:
  go test -race -v -coverprofile coverage.out ./...

show-cover:
  go tool cover -func coverage.out

smoke-test-cli msg='i cant go for that':
  go run cmd/ghoststring/main.go -help && \
    printf '%s' '{{ msg }}' | \
    go run cmd/ghoststring/main.go -k 'no can do' | \
    go run cmd/ghoststring/main.go -d -k 'no can do'

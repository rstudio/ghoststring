default: vet test

vet:
  go vet ./...

test:
  go test -v -coverprofile coverage.out ./...

show-cover:
  go tool cover -func coverage.out

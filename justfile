set shell := ['bash', '-ec']

goarch := `go env GOARCH 2>/dev/null || echo amd64`
goos := `go env GOOS 2>/dev/null || echo linux`
package := `awk '/^module/ { print $NF }' go.mod`

go_build_flags := '-v -ldflags "-extldflags \"-static\""'

default: vet test

vet:
  go vet ./...

clean:
  rm -rvf build/ .local/tmp/testlog.*

distclean: clean
  rm -rvf .local/tmp/

build: _build-bins

_build-bins: (_build-bin 'ghoststring') (_build-internal-bin 'rectangles') (_build-internal-bin 'myths')

_build-bin binname='ghoststring' _goos=goos _goarch=goarch:
  CGO_ENABLED=0 GOOS={{ _goos }} GOARCH={{ _goarch }} \
    go build \
      -o build/{{ _goos }}/{{ _goarch }}/{{ binname }} \
      {{ go_build_flags }} \
      ./cmd/{{ binname }}/

_build-internal-bin binname='ghoststring' _goos=goos _goarch=goarch:
  CGO_ENABLED=0 GOOS={{ _goos }} GOARCH={{ _goarch }} \
    go build \
      -o build/{{ _goos }}/{{ _goarch }}/{{ binname }} \
      {{ go_build_flags }} \
      ./internal/integration/cmd/{{ binname }}/

test: build
  go test -race -v -coverprofile coverage.out ./...

show-cover:
  go tool cover -func coverage.out

smoke-test-cli msg='i cant go for that':
  go run cmd/ghoststring/main.go -help && \
    printf '%s' '{{ msg }}' | \
    go run cmd/ghoststring/main.go -k 'no can do' | \
    go run cmd/ghoststring/main.go -d -k 'no can do'

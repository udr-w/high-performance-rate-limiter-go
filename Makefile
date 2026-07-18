.PHONY: test race vet bench verify

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

bench:
	go test -bench=. -benchmem ./...

verify: test race vet

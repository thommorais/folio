.PHONY: build check format test clean

BIN_DIR := bin
BINARY := folio
SRC := ./cmd/folio
GO ?= go
LDFLAGS := -s -w

build:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(SRC)

check:
	gofmt -l ./cmd ./internal
	$(GO) vet ./...
	$(GO) test ./...

format:
	gofmt -w ./cmd ./internal

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)

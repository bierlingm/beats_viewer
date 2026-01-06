.PHONY: build test vet release install clean

VERSION := $(shell grep 'const version' cmd/btv/main.go | cut -d'"' -f2)
BINARY := btv
BUILD_DIR := build

build:
	go build -o $(BINARY) ./cmd/btv/

install: build
	cp $(BINARY) ~/werk/bin/

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
	rm -rf $(BUILD_DIR)

# Release creates a git tag and builds the binary
release: vet build
	@echo "Releasing $(BINARY) v$(VERSION)"
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Tagged v$(VERSION). Push with: git push origin v$(VERSION)"

# Release a specific version (use: make release-version V=0.1.0)
release-version: vet
	@echo "Releasing $(BINARY) v$(V)"
	git tag -a "v$(V)" -m "Release v$(V)"
	@echo "Tagged v$(V). Push with: git push origin v$(V)"

# List all releases
releases:
	@git tag -l "v*" --sort=-version:refname

.PHONY: build test vet release install clean snapshot

VERSION := $(shell grep 'const version' cmd/btv/main.go | cut -d'"' -f2)
BINARY := btv

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
	rm -rf dist/

# Local snapshot build (no publish)
snapshot:
	goreleaser release --snapshot --clean

# Tag current version for release
tag: vet
	@echo "Tagging v$(VERSION)"
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Tagged v$(VERSION)"

# Tag a specific version (use: make tag-version V=0.1.0)
tag-version: vet
	@echo "Tagging v$(V)"
	git tag -a "v$(V)" -m "Release v$(V)"
	@echo "Tagged v$(V)"

# Full release via goreleaser (requires GITHUB_TOKEN)
release:
	goreleaser release --clean

# List all releases
releases:
	@git tag -l "v*" --sort=-version:refname

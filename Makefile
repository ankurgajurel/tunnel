VERSION ?=

.PHONY: test
test:
	go test ./...

.PHONY: build-cli
build-cli:
	mkdir -p dist
	go build -trimpath -ldflags="-s -w" -o dist/tunnel ./cmd/tunnel

.PHONY: build-release
build-release:
	rm -rf dist
	mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/tunnel-linux-amd64 ./cmd/tunnel
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/tunnel-linux-arm64 ./cmd/tunnel
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/tunnel-darwin-amd64 ./cmd/tunnel
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/tunnel-darwin-arm64 ./cmd/tunnel
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/tunnel-windows-amd64.exe ./cmd/tunnel

.PHONY: test-release
test-release: test build-release

.PHONY: release
release:
	@test -n "$(VERSION)" || (echo "usage: make release VERSION=v0.1.0" && exit 1)
	$(MAKE) test-release
	git diff --quiet
	git diff --cached --quiet
	git tag "$(VERSION)"
	git push origin master
	git push origin "$(VERSION)"
	@echo "release tag pushed: $(VERSION)"
	@echo "watch github actions to finish publishing binaries"

.PHONY: delete-release-tag
delete-release-tag:
	@test -n "$(VERSION)" || (echo "usage: make delete-release-tag VERSION=v0.1.0" && exit 1)
	git tag -d "$(VERSION)" || true
	git push origin ":refs/tags/$(VERSION)" || true

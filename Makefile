.PHONY: build test lint clean cover release update-consumers

build:
	GOWORK=off go build ./...
	cd redis && GOWORK=off go build ./...
	cd archive && GOWORK=off go build ./...

test:
	GOWORK=off go test -race -count=1 ./...

lint:
	GOWORK=off golangci-lint run ./...

cover:
	GOWORK=off go test -race -coverprofile=cover.out ./...
	go tool cover -func=cover.out | tail -1
	@rm -f cover.out

clean:
	go clean -cache -testcache

# Release: lint, test, tag all modules, push.
# Usage: make release V=0.4.0
release:
	@test -n "$(V)" || (echo "Usage: make release V=0.4.0" && exit 1)
	@echo "==> Lint + Test"
	$(MAKE) lint
	$(MAKE) test
	@echo "==> Tagging v$(V)"
	git add -A && git diff --cached --quiet || git commit -m "release v$(V)"
	git tag -a "v$(V)" -m "v$(V)"
	git tag -a "redis/v$(V)" -m "redis/v$(V)"
	git tag -a "archive/v$(V)" -m "archive/v$(V)"
	git push origin main "v$(V)" "redis/v$(V)" "archive/v$(V)"
	@echo "==> Done. Run: make update-consumers V=$(V)"

# Update all consumers to the new version.
# Usage: make update-consumers V=0.4.0
CONSUMERS := $(HOME)/src/go-hully $(HOME)/src/dozor $(HOME)/src/vaelor

update-consumers:
	@test -n "$(V)" || (echo "Usage: make update-consumers V=0.4.0" && exit 1)
	@for dir in $(CONSUMERS); do \
		name=$$(basename $$dir); \
		echo "==> Updating $$name"; \
		cd $$dir && \
		GOWORK=off go get github.com/anatolykoptev/go-session@v$(V) && \
		if [ -d vendor ]; then GOWORK=off go mod vendor; fi && \
		GOWORK=off go build ./... && \
		echo "    $$name: OK" || echo "    $$name: FAILED"; \
	done
	@echo "==> All consumers updated. Commit and deploy manually."

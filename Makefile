.DEFAULT_GOAL := generate

docs-prepare:
	cd docs && pnpm install

docs-debug: docs-prepare
	cd docs && pnpm dev

docs-build: docs-prepare
	cd docs && pnpm build

docs-preview: docs-build
	cd docs pnpm preview

up:
	cd docs && pnpm dlx @astrojs/upgrade
	go get -t -u ./... && go mod tidy
	cd _examples && go get -u ./... && go mod tidy

prepare:
	go mod tidy
	go generate -x ./...

examples: prepare
	cd _examples/ && go generate -x ./...
	cd _examples/ && go mod tidy
	cd _examples/kitchensink && go test -v -race -timeout 3m -count 2 ./...
	cd _examples/simple && go test -v -race -timeout 3m -count 2 ./...

dlv-kitchensink:
	cd _examples/kitchensink/internal && dlv debug \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		database/entc.go

ensure-clean:
	@if [ "${CI}" = "true" ] && [ "$(shell git status --porcelain | grep -v 'coverage')" != "" ]; then \
		echo "ERROR: git working directory is not clean. Make sure to run 'make test' and re-commit any changes."; \
		exit 1; \
	fi

test: prepare examples ensure-clean
	go test -v -race -timeout 3m -count 2 ./...

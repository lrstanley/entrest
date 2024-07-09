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
	cd _examples && go mod tidy

examples: prepare
	cd _examples/ && go generate -x ./...
	cd _examples/kitchensink && go test -v -race -timeout 3m -count 2 ./...
	cd _examples/simple && go test -v -race -timeout 3m -count 2 ./...

dlv-kitchensink:
	cd _examples/kitchensink/internal/database && dlv debug \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		generate.go

test: prepare examples
	go test -v -race -timeout 3m -count 2 ./...

.DEFAULT_GOAL := generate

docs-prepare:
	cd docs && pnpm install

docs-debug: docs-prepare
	cd docs && pnpm dev

docs-build: docs-prepare
	cd docs && pnpm build

docs-preview: docs-build
	cd docs && pnpm preview

up:
	$(eval SCALAR_VERSION=$(shell curl -sSq https://registry.npmjs.org/@scalar/api-reference/latest | jq -r '.version'))
	$(eval SCALAR_HASH=$(shell curl -sSq https://cdn.jsdelivr.net/npm/@scalar/api-reference@$(SCALAR_VERSION) | openssl dgst -sha256 -binary | openssl base64 -A))
	sed -ri -e "s:@scalar/api-reference@[0-9.]{5,10}:@scalar/api-reference@$(SCALAR_VERSION):g" -e 's:integrity="sha256-[^"]+":integrity="sha256-$(SCALAR_HASH)":g' templates/helper/server/docs.tmpl
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
	@if [ "${CI}" = "true" ] && [ "$(shell git status --porcelain | grep -Ev 'coverage|ghmeta')" != "" ]; then \
		echo "ERROR: git working directory is not clean. Make sure to run 'make test' and re-commit any changes."; \
		git status --porcelain | grep -Ev 'coverage|ghmeta'; \
		exit 1; \
	fi

test: prepare examples ensure-clean
	go test -v -race -timeout 3m -count 2 ./...

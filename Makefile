.DEFAULT_GOAL := generate

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

up:
	go get -t -u ./... && go mod tidy
	cd _examples && go get -u ./... && go mod tidy

prepare:
	go mod tidy
	cd _examples && go mod tidy

simple: prepare
	cd _examples/simple && go generate -x ./...

debug: prepare
	cd _examples/debug && go generate -x ./...

dlv-debug:
	cd _examples/debug && dlv debug \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		generate.go

test: prepare
	go test -v -race -timeout 3m -count 2 ./...

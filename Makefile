.DEFAULT_GOAL := generate

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

up:
	go get -t -u ./... && go mod tidy
	cd _examples && go get -u ./... && go mod tidy

prepare:
	go mod tidy
	cd _examples && go mod tidy

kitchensink: prepare
	cd _examples/kitchensink && go generate -x ./...

dlv-kitchensink:
	cd _examples/kitchensink && dlv kitchensink \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		generate.go

test: prepare
	go test -v -race -timeout 3m -count 2 ./...

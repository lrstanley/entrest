.DEFAULT_GOAL := generate

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

up: go-upgrade-deps
	@echo

go-fetch:
	go mod tidy

go-upgrade-deps:
	go get -u ./... && go mod tidy

simple:
	cd _examples/simple && go generate -x ./...

dlv-simple:
	cd _examples/simple && dlv debug \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		generate.go

clean:
	go clean -r -cache -testcache -modcache
.PHONY: clean

tidy:
	go mod tidy -v -x
.PHONY: tidy

ifndef app_version
app_version := dev
endif
build:
	rm -rfv ./bin
	mkdir -vp ./bin
	go build -tags urfave_cli_no_docs -trimpath -buildvcs=false -ldflags "-extldflags '-static' -s -w -buildid='' -X 'main.AppVersion=${app_version}' -X 'main.AppCompileTime=$(shell date --iso-8601=seconds)' -X 'main.AppName=fx-news'" -o ./bin/ingest .
.PHONY: build

build-debug:
	rm -rfv ./bin
	mkdir -vp ./bin
	go build -tags urfave_cli_no_docs -buildvcs=false -ldflags "-compressdwarf=false -extldflags '-static' -buildid='' -X 'main.AppVersion=${app_version}' -X 'main.AppCompileTime=$(shell date --iso-8601=seconds)' -X 'main.AppName=fx-news'" -o ./bin/ingest .
.PHONY: build-debug

build-clean: clean build
.PHONY: build-clean

test:
	go test -trimpath -buildvcs=false -ldflags '-extldflags "-static" -s -w -buildid=' -race -failfast -vet=all -covermode=atomic -coverprofile=coverage.out -v ./...
.PHONY: test

outdated-indirect:
	go list -u -m -f '{{if and .Update .Indirect}}{{.}}{{end}}' all
.PHONY: outdated-indirect

outdated-direct:
	go list -u -m -f '{{if and .Update (not .Indirect)}}{{.}}{{end}}' all
.PHONY: outdated-direct

outdated-all: outdated-direct outdated-indirect
.PHONY: outdated-all

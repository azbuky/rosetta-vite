.PHONY: deps build run lint run-mainnet-online run-mainnet-offline run-testnet-online \
	run-testnet-offline check-comments shorten-lines \
	spellcheck salus build-local format check-format test coverage coverage-local \
	update-bootstrap-balances mocks

SPELLCHECK_CMD=go run github.com/client9/misspell/cmd/misspell
GOLINES_CMD=go run github.com/segmentio/golines
GOLINT_CMD=go run golang.org/x/lint/golint
GOVERALLS_CMD=go run github.com/mattn/goveralls
GOIMPORTS_CMD=go run golang.org/x/tools/cmd/goimports
GO_PACKAGES=./services/... ./cmd/... ./configuration/... ./vite/... 
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")
TEST_SCRIPT=go test ${GO_PACKAGES}
LINT_SETTINGS=golint,misspell,gocyclo,gocritic,whitespace,goconst,gocognit,bodyclose,unconvert,lll,unparam
PWD=$(shell pwd)
NOFILE=102400

deps:
	go get ./...

test:
	${TEST_SCRIPT}

build:
	docker build -t rosetta-vite:latest https://github.com/azbuky/rosetta-vite.git

build-local:
	docker build -t rosetta-vite:latest .

build-release:
	# make sure to always set version with vX.X.X
	docker build -t rosetta-vite:$(version) .;
	docker save rosetta-vite:$(version) | gzip > rosetta-vite-$(version).tar.gz;

update-bootstrap-balances:
	go run main.go utils:generate-bootstrap vite/genesis_files/mainnet.json rosetta-cli-conf/mainnet/bootstrap_balances.json;
	
run-mainnet-online:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -v "${PWD}/vite-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-vite:latest

run-mainnet-offline:
	docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-vite:latest

run-testnet-online:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -v "${PWD}/vite-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-vite:latest

run-testnet-offline:
	docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-vite:latest

run-mainnet-remote:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -e "GVITE=$(gvite)" -p 8080:8080 -p 30303:30303 rosetta-vite:latest

run-testnet-remote:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -e "GVITE=$(gvite)" -p 8080:8080 -p 30303:30303 rosetta-vite:latest

check-comments:
	${GOLINT_CMD} -set_exit_status ${GO_FOLDERS} .

lint: | check-comments
	golangci-lint run --timeout 2m0s -v -E ${LINT_SETTINGS},gomnd

shorten-lines:
	${GOLINES_CMD} -w --shorten-comments ${GO_FOLDERS} .

format:
	gofmt -s -w -l .
	${GOIMPORTS_CMD} -w .

check-format:
	! gofmt -s -l . | read
	! ${GOIMPORTS_CMD} -l . | read

salus:
	docker run --rm -t -v ${PWD}:/home/repo coinbase/salus

spellcheck:
	${SPELLCHECK_CMD} -error .

coverage:	
	if [ "${COVERALLS_TOKEN}" ]; then ${TEST_SCRIPT} -coverprofile=c.out -covermode=count; ${GOVERALLS_CMD} -coverprofile=c.out -repotoken ${COVERALLS_TOKEN}; fi

coverage-local:
	${TEST_SCRIPT} -cover

mocks:
	rm -rf mocks;
	mockery --dir services --all --case underscore --outpkg services --output mocks/services;
	mockery --dir vite --all --case underscore --outpkg vite --output mocks/vite;

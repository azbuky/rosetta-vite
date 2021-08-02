<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Rosetta" src="https://www.rosetta-api.org/img/rosetta_header.png">
  </a>
</p>
<h3 align="center">
   Rosetta Vite
</h3>

<p align="center"><b>
ROSETTA-VITE IS CONSIDERED <a href="https://en.wikipedia.org/wiki/Software_release_life_cycle#Alpha">ALPHA SOFTWARE</a>.
USE AT YOUR OWN RISK! THE AUTHORS ASSUME NO RESPONSIBILITY NOR LIABILITY IF THERE IS A BUG IN THIS IMPLEMENTATION.
</b></p>

## Overview

`rosetta-vite` provides a implementation of the Rosetta API for
Vite in Golang. If you haven't heard of the Rosetta API, you can find more
information [here](https://rosetta-api.org).

## Features

* Full support for rosetta data and construction apis

## Usage

As specified in the [Rosetta API Principles](https://www.rosetta-api.org/docs/automated_deployment.html),
all Rosetta implementations must be deployable via Docker and support running via either an
[`online` or `offline` mode](https://www.rosetta-api.org/docs/node_deployment.html#multiple-modes).

**YOU MUST INSTALL DOCKER FOR THE FOLLOWING INSTRUCTIONS TO WORK. YOU CAN DOWNLOAD
DOCKER [HERE](https://www.docker.com/get-started).**

### Install

Running the following commands will create a Docker image called `rosetta-vite:latest`.

#### From Source

After cloning this repository, run:

```text
make build-local
```

### Run

Running the following commands will start a Docker container in
[detached mode](https://docs.docker.com/engine/reference/run/#detached--d) with
a data directory at `<working directory>/vite-data` and the Rosetta API accessible
at port `8080`.

#### Configuration Environment Variables

* `MODE` (required) - Determines if Rosetta can make outbound connections. Options: `ONLINE` or `OFFLINE`.
* `NETWORK` (required) - Vite network to launch and/or communicate with. Options: `MAINNET`, `TESTNET`.
* `PORT`(required) - Which port to use for Rosetta.
* `GVITE` (optional) - Point to a remote `gvite` node instead of initializing one

#### Mainnet:Online

```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/vite-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-vite:latest
```

_If you cloned the repository, you can run `make run-mainnet-online`._

#### Mainnet:Online (Remote)

```text
docker run -d --rm --ulimit "nofile=100000:100000" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -e "GVITE=<NODE URL>" -p 8080:8080 -p 30303:30303 rosetta-vite:latest
```

_If you cloned the repository, you can run `make run-mainnet-remote gvite=<NODE URL>`._

#### Mainnet:Offline

```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-vite:latest
```

_If you cloned the repository, you can run `make run-mainnet-offline`._

#### Testnet:Online

```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/vite-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-vite:latest
```

_If you cloned the repository, you can run `make run-testnet-online`._

#### Testnet:Online (Remote)

```text
docker run -d --rm --ulimit "nofile=100000:100000" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -e "GVITE=<NODE URL>" -p 8080:8080 -p 30303:30303 rosetta-vite:latest
```

_If you cloned the repository, you can run `make run-testnet-remote gvite=<NODE URL>`._

#### Testnet:Offline

```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-gvite:latest
```

_If you cloned the repository, you can run `make run-testnet-offline`._

## Testing with rosetta-cli

To validate `rosetta-vite`, [install `rosetta-cli`](https://github.com/coinbase/rosetta-cli#install)
and run one of the following commands:

* `rosetta-cli check:data --configuration-file rosetta-cli-conf/testnet/config.json`
* `rosetta-cli check:construction --configuration-file rosetta-cli-conf/testnet/config.json`
* `rosetta-cli check:data --configuration-file rosetta-cli-conf/mainnet/config.json`

## Development

* `make deps` to install dependencies
* `make test` to run tests
* `make lint` to lint the source code
* `make salus` to check for security concerns
* `make build-local` to build a Docker image from the local context
* `make coverage-local` to generate a coverage report

## License

This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).

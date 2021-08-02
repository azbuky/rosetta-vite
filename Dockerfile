
# Compile golang 
FROM ubuntu:18.04 as golang-builder

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

RUN apt-get update && apt-get install -y curl make gcc g++ git
ENV GOLANG_VERSION 1.15.5
ENV GOLANG_DOWNLOAD_SHA256 9a58494e8da722c3aef248c9227b0e9c528c7318309827780f16220998180a0d
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Compile gvite
FROM golang-builder as gvite-builder

RUN go get github.com/vitelabs/go-vite \
  && cd $GOPATH/src/github.com/vitelabs/go-vite \
  && make gvite

RUN cd $GOPATH/src/github.com/vitelabs/go-vite \
  && mv build/cmd/gvite/gvite /app/gvite \
  && rm -rf go-vite

# Compile rosetta-vite
FROM golang-builder as rosetta-builder

# Use native remote build context to build in any directory
COPY . src 
RUN cd src \
  && go build

RUN mv src/rosetta-vite /app/rosetta-vite \
  && mkdir /app/vite \
  && mv src/node_config.json /app/vite/node_config.json \
  && rm -rf src 

## Build Final Image
FROM ubuntu:18.04

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app \
  && mkdir -p /data \
  && chown -R nobody:nogroup /data

WORKDIR /app

# Copy binary from gvite-builder
COPY --from=gvite-builder /app/gvite /app/gvite

# Copy binary from rosetta-builder
COPY --from=rosetta-builder /app/vite /app/vite
COPY --from=rosetta-builder /app/rosetta-vite /app/rosetta-vite

# Set permissions for everything added to /app
RUN chmod -R 755 /app/*

CMD ["/app/rosetta-vite", "run"]

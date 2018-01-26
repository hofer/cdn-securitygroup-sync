FROM ubuntu:16.04

RUN apt-get update -y && apt-get install -y wget curl rake git

# Golang
RUN wget -O /usr/local/go1.9.linux-amd64.tar.gz https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz && \
    cd /usr/local/ && tar -xvzf go1.9.linux-amd64.tar.gz && \
    rm -rf go1.9.linux-amd64.tar.gz && \
    mkdir -p /usr/local/go_work

ENV GOROOT /usr/local/go
ENV GOPATH /usr/local/go_work
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

RUN go get github.com/jteeuwen/go-bindata/...
RUN go get github.com/elazarl/go-bindata-assetfs/...

# Glide
RUN curl https://glide.sh/get | sh

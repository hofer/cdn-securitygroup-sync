VERSION ?= $(shell git describe --tags | cut -dv -f2)
LDFLAGS := -X main.AppVersion=$(VERSION)
HANDLER ?= main
PACKAGE ?= $(HANDLER)
CURDIR := $(shell pwd)
IMG_NAME := cd-securitygroup-sync:latest
WORKDIR := /usr/local/go/src/cdn-securitygroup-sync/
RUN_CMD := docker run --rm=true -v $(CURDIR):$(WORKDIR) -w $(WORKDIR) $(IMG_NAME)

all: dependencies build

zip: all pack

dependencies:
	docker build -t cd-securitygroup-sync:latest .
	$(RUN_CMD) glide install

build:
	$(RUN_CMD) env GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(HANDLER) .

pack:
	$(RUN_CMD) zip $(PACKAGE).zip $(HANDLER)

clean:
	$(RUN_CMD) rm -rf $(HANDLER) $(PACKAGE).zip vendor glide.lock

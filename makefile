VERSION := $(shell git describe --tags --dirty --always)
SERVICE := $(shell basename $(shell pwd))
OWNER := southclaws
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
-include .env

# -
# Local Development
#-

static:
	go get
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o $(SERVICE) .

fast:
	go build $(LDFLAGS) -o $(SERVICE)

local: fast
	./$(SERVICE)

test:
	go test -v -race ./storage
	go test -v -race ./bot/commands

version:
	git tag $(VERSION)
	git push
	git push origin $(VERSION)


# -
# Docker
# -

build:
	docker build --no-cache -t $(OWNER)/$(SERVICE):$(VERSION) .

push:
	docker push $(OWNER)/$(SERVICE):$(VERSION)
	docker tag $(OWNER)/$(SERVICE):$(VERSION) $(OWNER)/$(SERVICE):latest
	docker push $(OWNER)/$(SERVICE):latest

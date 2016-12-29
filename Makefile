VERSION := 1.2.1
DOCKER_TAG := haproxy-docker-wrapper:$(VERSION)_1.6.2
PACKAGE := github.com/tuenti/haproxy-docker-wrapper
ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all:
	docker run -v $(ROOT_DIR):/go/src/$(PACKAGE) -w /go/src/$(PACKAGE) -it --rm golang:1.7.4 go build -ldflags "-X main.version=$(VERSION)"

release:
	@if echo $(VERSION) | grep -q "dev$$" ; then echo Set VERSION variable to release; exit 1; fi
	@if git show v$(VERSION) > /dev/null 2>&1; then echo Version $(VERSION) already exists; exit 1; fi
	sed -i "s/^VERSION :=.*/VERSION := $(VERSION)/" Makefile
	git ci Makefile -m "Version $(VERSION)"
	git tag v$(VERSION) -a -m "Version $(VERSION)"

docker: all
	docker build -t $(DOCKER_TAG) .

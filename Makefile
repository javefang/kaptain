# variables
BASE_PKG        := github.com/javefang/kaptain

VERSION         ?= $(shell git describe --tags | sort | head -1)
COMMIT          ?= $(shell git rev-parse HEAD)
TREESTATE       ?= $(shell [ -z "$$(git status --porcelain)" ] && echo "clean" || echo "dirty")

BUILD_DIR       ?= build
BUILD_IMAGE     ?= golang:1.10.0

# go option
GO      ?= go
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)
LDFLAGS := -X $(BASE_PKG)/pkg/version.version=$(VERSION) -X $(BASE_PKG)/pkg/version.gitCommit=$(COMMIT) -X $(BASE_PKG)/pkg/version.gitTreeState=$(TREESTATE)

# Build and install the binary
.PHONY: install
install: vendor data/assets.go
	$(GO) install -ldflags '$(LDFLAGS)' $(BASE_PKG)/cmd/kaptain
	$(GO) install -ldflags '$(LDFLAGS)' $(BASE_PKG)/cmd/sailor

.PHONY: acc-test
acc-test: install
	scripts/acctest.sh

.PHONY: release
release: $(BUILD_DIR)/linux-amd64/kaptain $(BUILD_DIR)/darwin-amd64/kaptain $(BUILD_DIR)/windows-amd64/kaptain $(BUILD_DIR)/linux-amd64/sailor

.PHONY: docker-build
docker-build:
	docker run --rm -v $(shell pwd):/go/src/$(BASE_PKG) -w /go/src/$(BASE_PKG) $(BUILD_IMAGE) go get -u github.com/jteeuwen/go-bindata/... && make release

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	
### File Targets ###

$(BUILD_DIR)/linux-amd64/kaptain: vendor data/assets.go
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/linux-amd64/kaptain $(BASE_PKG)/cmd/kaptain

$(BUILD_DIR)/darwin-amd64/kaptain: vendor data/assets.go
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/darwin-amd64/kaptain $(BASE_PKG)/cmd/kaptain

$(BUILD_DIR)/windows-amd64/kaptain: vendor data/assets.go
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/windows-amd64/kaptain $(BASE_PKG)/cmd/kaptain

$(BUILD_DIR)/linux-amd64/sailor: vendor
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/linux-amd64/sailor $(BASE_PKG)/cmd/sailor

data/assets.go: bin/go-bindata $(shell find assets -type f)
	./bin/go-bindata -o data/assets.go -pkg data assets/...

vendor: bin/dep Gopkg.lock
	./bin/dep ensure
	
# Tools

bin/dep:
	@echo "Downloading dep"
	mkdir -p bin
	curl -o bin/dep -L https://github.com/golang/dep/releases/download/v0.3.2/dep-$(GOOS)-$(GOARCH)
	chmod +x bin/dep
	
bin/go-bindata:
	@echo "Downloading go-bindata"
	mkdir -p bin
	$(GO) get -u github.com/jteeuwen/go-bindata/...
	cp $$GOPATH/bin/go-bindata bin/

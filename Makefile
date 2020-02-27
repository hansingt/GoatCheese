IMAGE_NAME ?= GoatCheese
IMAGE_TAG ?= dev
COVERAGE_FILE := .coverage
SRC_FILES := $(wildcard ./**/*.go)
EXECUTABLE := ./GoatCheese

MODULE := github.com/hansingt/GoatCheese

.PHONY: all \
		build \
		cover cover-show cover-html \
		deps \
		image \
		lint \
		run \
		test

all: cover

build: $(EXECUTABLE)
$(EXECUTABLE): $(SRC_FILES) | deps
	go build -ldflags "-s -w" -o $@ $(MODULE)/cmd/GoatCheese

run: $(EXECUTABLE)
	$(EXECUTABLE) $(RUNOPTS)

test: $(SRC_FILES) | deps
	go test -race $(GOTEST_OPTS) ./...

cover: $(COVERAGE_FILE)
$(COVERAGE_FILE): $(SRC_FILES) | deps
	$(MAKE) test GOTEST_OPTS="-coverprofile=$(COVERAGE_FILE)"

cover-show: $(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE)

cover-html: $(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html

lint:
	docker run --rm -it \
		-v $(PWD):/go/src/$(MODULE):ro \
		-w /go/src/$(MODULE) \
		-e GO11MODULE="on" \
		golangci/golangci-lint:latest-alpine \
		golangci-lint run

image: .image
.image: $(SRC_FILES) Dockerfile templates
	docker build --force-rm --pull -t $(IMAGE_NAME):$(IMAGE_TAG) .

IMAGE_NAME ?= pypigo
IMAGE_TAG ?= dev

COVERAGE_FILE := .coverage
SRC_FILES := $(wildcard ./**/*.go)
EXECUTABLE := ./PyPiGo

.PHONY: build \
		cover cover-show cover-html \
		image \
		run \
		test


build: $(EXECUTABLE)
$(EXECUTABLE): $(SRC_FILES)
	go build -o $@ .

run: $(EXECUTABLE)
	$(EXECUTABLE) $(RUNOPTS)

test: $(SRC_FILES)
	go test -v $(GOTEST_OPTS) ./...

cover: $(COVERAGE_FILE)
$(COVERAGE_FILE): $(SRC_FILES)
	$(MAKE) test GOTEST_OPTS="-coverprofile=$(COVERAGE_FILE)"

cover-show:
	go tool cover -html=$(COVERAGE_FILE)

cover-html: $(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html

image: .image
.image: $(SRC_FILES) Dockerfile templates
	docker build --force-rm --pull -t $(IMAGE_NAME):$(IMAGE_TAG) .

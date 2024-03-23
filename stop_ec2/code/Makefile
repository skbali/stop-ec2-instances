
# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./main.go
BINARY_NAME := bootstrap
BUILD_DIR := build

.DEFAULT_GOAL := all

.PHONE: all
all: clean tidy audit lint test mac


.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: clean
clean:
	rm -rf ${BUILD_DIR}

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


.PHONY: test
test:
	go test -v ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -coverprofile=./coverage.out ./...
	go tool cover -html=./coverage.out

## lint: run linters
.PHONY: lint
lint:
	golangci-lint run ./...

## build: build the application
.PHONY: build
build:
	@make --no-print-directory build-platform GOOS=linux GOARCH=amd64 CGO_ENABLED=0

.PHONY: win
win:
	@make --no-print-directory build-platform GOOS=windows GOARCH=amd64 CGO_ENABLED=0

.PHONY: mac
mac:
	@make --no-print-directory build-platform GOOS=darwin GOARCH=arm64 CGO_ENABLED=0

.PHONY: build-platform
build-platform: clean
	@echo "Building for $(GOOS)/$(GOARCH)"
	@mkdir -p ${BUILD_DIR}
	$(eval BINARY_NAME := ${BINARY_NAME}$(if $(findstring windows,$(GOOS)),.exe,))
	@go build -v -o ${BUILD_DIR}/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

all: compile upload

compile: check-go build zip
upload:compile to-s3 to-lambda clean-all

# defn for default path and args for build command
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0
BUILD_DIR ?= ./build
MAIN_DIR ?= ./function

GO_FILES ?= $(shell find . -name '*.go')

BUILD_FILE ?= main

check-go:
	@which go > /dev/null || (echo "Go not found. Please install Go." && exit 1)

build:
	@go env -w GOOS=$(GOOS)
	@go env -w GOARCH=$(GOARCH)
	@go env -w CGO_ENABLED=$(CGO_ENABLED)
	go build -o $(BUILD_DIR)/${BUILD_FILE} $(GO_FILES)

zip: 
	@which zip > /dev/null || (echo "zip not found. Please install zip." && exit 1)
	zip -FSr $(BUILD_DIR)/${BUILD_FILE}.zip $(BUILD_DIR)/${BUILD_FILE}


BUCKET_NAME ?= shoestring-lambda-bucket

to-s3:
	aws s3 sync $(BUILD_DIR)/ s3://${BUCKET_NAME} --exclude "*" --include "*.zip"

FUNC_NAME ?= update-json 
BUCKET_NAME ?= shoestring-lambda-bucket

to-lambda:
	aws lambda update-function-code --function-name ${FUNC_NAME} --s3-bucket ${BUCKET_NAME} --s3-key ${BUILD_FILE}.zip --no-cli-pager

clean-all:
	rm -rf $(BUILD_DIR)

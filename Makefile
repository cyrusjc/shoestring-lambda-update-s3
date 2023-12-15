all: build zip clean

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./build/main ./function/main.go

zip: 
	zip ./build/main.zip ./build/main

clean:
	rm ./build/main
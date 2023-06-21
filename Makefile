BINARY_NAME=s9k

all: build

build:
		go build -o bin/${BINARY_NAME} cmd/${BINARY_NAME}/${BINARY_NAME}.go
 
run:
		go run cmd/${BINARY_NAME}/${BINARY_NAME}.go -debug

install:
		go install cmd/${BINARY_NAME}/${BINARY_NAME}.go

clean:
		go clean
		rm -rf bin/${BINARY_NAME}

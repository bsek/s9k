BINARY_NAME=s9k

all: build

build:
		go build -o bin/${BINARY_NAME} cmd/s9k/s9k.go
 
run:
		go run cmd/s9k/s9k.go

install:
		go install cmd/s9k/s9k.go

clean:
		go clean
		rm -rf bin/${BINARY_NAME}

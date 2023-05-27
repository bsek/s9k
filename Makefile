BINARY_NAME=s9k

all: build

build:
		go build -o bin/${BINARY_NAME} cmd/s9k/main.go
 
run:
		go run cmd/s9k/main.go

install:
		go install -o s9k cmd/s9k/main.go

clean:
		go clean
		rm -rf bin/${BINARY_NAME}

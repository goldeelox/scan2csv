# https://just.systems
default: fmt vet

fmt:
  go fmt ./...

vet:
  go vet -v ./...

build:
  go build -ldflags "-X main.Version=$(git describe --tags --abbrev=0) -X main.Revision=$(git rev-parse HEAD)" -v -o ./dist/ ./...

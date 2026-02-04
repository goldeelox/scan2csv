# https://just.systems
default: fmt vet

fmt:
  go fmt ./...

vet:
  go vet -v ./...

build:
  CGO_ENABLED=0 go build -ldflags "-X main.Version=$(git describe --tags --abbrev=0) -X main.Revision=$(git rev-parse HEAD)" -v -o ./dist/ ./...

run:
  rm -rf scans*.csv
  go run ./...
  cat scans*.csv

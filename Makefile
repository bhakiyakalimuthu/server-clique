all: clean build test test-race lint gofumpt docker-image docker-run
APP_NAME := server-clique

GOPATH := $(if $(GOPATH),$(GOPATH),~/go)
VERSION := $(shell git describe --tags --always)

clean:
	rm -rf ${APP_NAME}-* *.json build/

build:
	go build -trimpath -ldflags "-X main._BuildVersion=${VERSION}" -v -o ${APP_NAME}-server cmd/server/main.go
	go build -trimpath -ldflags "-X main._BuildVersion=${VERSION}" -v -o ${APP_NAME}-client cmd/client/main.go

test:
	go test ./...

test-race:
	go test -race ./...

mod:
	go mod tidy

lint:
	gofmt -d -s .
	gofumpt -d -extra .
	go vet ./...
	staticcheck ./...
	golangci-lint run

gofumpt:
	gofumpt -l -w -extra .


# server
docker-image-server:
	DOCKER_BUILDKIT=1 docker build --platform linux/amd64 --progress=plain  --build-arg VERSION=${VERSION} -f dockerfiles/server/Dockerfile . -t ${APP_NAME}-server-${VERSION}:${VERSION}

osx-docker-image-server:
	DOCKER_BUILDKIT=1 docker build --platform linux/arm64  --progress=plain  --build-arg APP_NAME=${APP_NAME}-server --build-arg VERSION=${VERSION} -f dockerfiles/server/Dockerfile . -t ${APP_NAME}-server-${VERSION}:latest

docker-run-server:
	docker run --network=host  ${APP_NAME}-server-${VERSION}


# client
docker-image-client:
	DOCKER_BUILDKIT=1 docker build --platform linux/amd64 --progress=plain  --build-arg VERSION=${VERSION} -f dockerfiles/client/Dockerfile . -t ${APP_NAME}-client-${VERSION}:${VERSION}

osx-docker-image-client:
	DOCKER_BUILDKIT=1 docker build --platform linux/arm64  --progress=plain  --build-arg APP_NAME=${APP_NAME}-client --build-arg VERSION=${VERSION} -f dockerfiles/client/Dockerfile . -t ${APP_NAME}-client-${VERSION}:latest

docker-run-client:
	docker run  --network=host  ${APP_NAME}-client-${VERSION}


# queue
docker-run-rabbitmq:
	docker run --rm -it -p 15672:15672 -p 5672:5672 rabbitmq:3-management

APP=vault-token-renewer
REGISTRY?=rtakaishi
COMMIT_SHA=$(shell git rev-parse --short HEAD)

.PHONY: build
## build: build the application
build: clean
	go build -o ${APP} main.go

.PHONY: run
## run: runs go run main.go
run:
	go run -race main.go

.PHONY: clean
## clean: cleans the binary
clean:
	go clean

.PHONY: docker-build
## docker-build: build container image
docker-build:
	echo docker build -t ${APP} .
	echo docker tag ${APP} ${REGISTRY}/${APP}:${COMMIT_SHA}

.PHONY: docker-push
## docker-push: push container image
docker-push: docker-build
	docker push ${REGISTRY}/${APP}:${COMMIT_SHA}

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

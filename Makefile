# Definitions
ROOT                    := $(PWD)
GO_HTML_COV             := ./coverage.html
GO_TEST_OUTFILE         := ./c.out
GOLANG_DOCKER_IMAGE     := golang:1.15
GOLANG_DOCKER_CONTAINER := goesquerydsl-container

#   Format according to gofmt: https://github.com/cytopia/docker-gofmt
#   Usage:
#       make fmt
#       make fmt path=src/elastic/index_setup.go
fmt:
ifdef path
	docker run --rm -v ${ROOT}:/data cytopia/gofmt -s -w ${path}
else
	docker run --rm -v ${ROOT}:/data cytopia/gofmt -s -w .
endif

#   Deletes container if exists
#   Usage:
#       make clean
clean:
	docker rm -f ${GOLANG_DOCKER_CONTAINER} || true

#   Usage:
#       make test
test:
	docker run -w /app -v ${ROOT}:/app ${GOLANG_DOCKER_IMAGE} go test ./... -coverprofile=${GO_TEST_OUTFILE}
	docker run -w /app -v ${ROOT}:/app ${GOLANG_DOCKER_IMAGE} go tool cover -html=${GO_TEST_OUTFILE} -o ${GO_HTML_COV}

#   Usage:
#       make lint
lint:
	docker run --rm -v ${ROOT}:/data cytopia/gofmt -s -w .
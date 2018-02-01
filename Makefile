BIN=wallet
BUILD_OUTPUT=.
GO=go

build:
	${GO} build -o ${BUILD_OUTPUT}/${BIN} .

lint:
	${GO} get github.com/golang/lint/golint
	${GO} vet ./...
	golint ./...

test:
	${GO} test -v ./...

deps:
	${GO} get github.com/mr-tron/base58/base58
	${GO} get github.com/spaolacci/murmur3
	${GO} get -d github.com/toxeus/go-secp256k1 && \
	cd ${GOPATH}/src/github.com/toxeus/go-secp256k1 && \
	git submodule update --init && \
	cd c-secp256k1 && \
	./autogen.sh && ./configure && make && \
	cd .. && \
	go install

clean:
	rm -f $(BIN)
	${GO} clean

.PHONY: test build lint deps clean

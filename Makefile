BIN="gorond"

all: clean test build

test:
	go test ./...

build: deps
	go build -o build/$(BIN)

run: build
	./build/$(BIN)

deps:
	go get -d -t -v ./...
	go get github.com/aws/aws-sdk-go/service/sns
	go get github.com/robfig/cron
	go get gopkg.in/gcfg.v1

rpm:
	cp rpmbuild/goron.sample.conf rpmbuild/src/gorond.conf
	rpmbuild --define "_sourcedir `pwd`/rpmbuild/src" --define "_builddir `pwd`/build" --define "_logdir /var/log" -ba rpmbuild/gorond.spec

clean:
	rm -rf build/$(BIN)
	go clean

.PHONY: test build deps rpm clean


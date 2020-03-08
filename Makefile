.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/create cmd/create/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get cmd/get/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

test:
	go test ./... -v

deploy: clean build test
	sls deploy --verbose

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

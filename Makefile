.DEFAULT_GOAL := build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	export GO111MODULE=on
	env CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/sendsms SendSMS/main.go
	env CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/verify Verify/main.go

clean:
	rm -rf ./bin ./vendor

deploy: clean build
	sls deploy --verbose

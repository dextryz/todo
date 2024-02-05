fmt:
	go mod tidy -compat=1.17
	gofmt -l -s -w .

run:
	go run .

build:
	go build .

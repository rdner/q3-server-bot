SHELL=/bin/bash

bin/q3serverbot:
	go build -o ./bin/q3serverbot \
    -v -ldflags "\
    -extldflags \"-static\"" \
    .

.PHONY: test
test:
	go test ./... -coverprofile=coverage.txt -cover -v -timeout 10s


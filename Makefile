SHELL=/bin/bash

bin/q3serverbot:
	go build -o ./bin/q3serverbot \
    -v -ldflags "\
    -extldflags \"-static\"" \
    .

.PHONY: test
test:
	go test ./... -coverprofile=coverage.txt -cover -v -timeout 10s

clean:
	rm ./bin/q3serverbot
	rm ./q3serverbot.service

./q3serverbot.service:
	./unit.sh ./q3serverbot.service

.PHONY: systemd
systemd: ./q3serverbot.service

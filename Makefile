BINARY := spektacular
VERSION := 0.1.0

.PHONY: build test lint clean install install-local cross harbor-test

build:
	go build -ldflags "-X github.com/jumppad-labs/spektacular/cmd.version=$(VERSION)" -o $(BINARY) .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY)-*

install: build
	cp $(BINARY) $(GOPATH)/bin/$(BINARY)

install-local: build
	sudo cp $(BINARY) /usr/local/bin/$(BINARY)

cross:
	GOOS=darwin  GOARCH=arm64 go build -o $(BINARY)-darwin-arm64  .
	GOOS=darwin  GOARCH=amd64 go build -o $(BINARY)-darwin-amd64  .
	GOOS=linux   GOARCH=amd64 go build -o $(BINARY)-linux-amd64   .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY)-windows-amd64.exe .

harbor-test:
	GOOS=linux GOARCH=amd64 go build -o tests/harbor/spec-workflow/environment/spektacular .
	ANTHROPIC_AUTH_TOKEN=$$(python3 -c "import json; print(json.load(open('$$HOME/.claude/.credentials.json'))['claudeAiOauth']['accessToken'])") \
	harbor run -p tests/harbor/spec-workflow -a claude-code -m claude-sonnet-4-6 -o tests/harbor/jobs

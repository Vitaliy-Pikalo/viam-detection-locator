BIN := bin/detection-locator

.PHONY: test build clean

test:
	go test ./...

$(BIN): go.mod $(shell find . -name '*.go')
	go mod tidy
	go build -o $(BIN) ./cmd/module

build: $(BIN)

module.tar.gz: $(BIN) meta.json
	tar czf $@ meta.json $(BIN)

clean:
	rm -rf bin module.tar.gz

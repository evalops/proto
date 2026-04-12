.PHONY: lint breaking generate test clean

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'

generate:
	buf generate

test:
	go test ./...

clean:
	rm -rf gen/go gen/ts gen/python
	mkdir -p gen/go gen/ts gen/python

.PHONY: lint breaking generate clean

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'

generate:
	buf generate

clean:
	rm -rf gen/go gen/ts gen/python
	mkdir -p gen/go gen/ts gen/python

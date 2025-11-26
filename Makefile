.PHONY: hurl
hurl:
	hurl --test --jobs 1 hurl/*.hurl

build:
	go build ./...
test:
	go test -v -race -count 1 -cover ./...

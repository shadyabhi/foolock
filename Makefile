.PHONY: hurl hurl-test server-start server-stop

server-start:
	@go build -o /tmp/foolock .
	@rm -f /tmp/foolock.log
	@/tmp/foolock > /tmp/foolock.log 2>&1 & echo $$! > /tmp/foolock.pid
	@while ! curl -s http://localhost:8080/lock > /dev/null 2>&1; do sleep 0.1; done
	@echo "Server started with PID $$(cat /tmp/foolock.pid)"

hurl-test:
	hurl --test --jobs 1 hurl/*.hurl

server-stop:
	@kill $$(cat /tmp/foolock.pid) 2>/dev/null || true
	@rm -f /tmp/foolock.pid
	@echo "Server stopped"

hurl: server-start hurl-test server-stop

build:
	go build ./...
test:
	go test -v -race -count 1 -cover ./...

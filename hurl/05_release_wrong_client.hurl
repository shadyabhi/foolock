POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200

DELETE http://localhost:8080/lock?client=laptop2
HTTP 403
[Asserts]
jsonpath "$.error" == "client does not hold the lock"

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200

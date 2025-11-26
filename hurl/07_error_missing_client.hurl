POST http://localhost:8080/lock
HTTP 400
[Asserts]
jsonpath "$.error" == "client parameter required"

DELETE http://localhost:8080/lock
HTTP 400
[Asserts]
jsonpath "$.error" == "client parameter required"

POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200
[Asserts]
jsonpath "$.status" == "lock released"

GET http://localhost:8080/lock
HTTP 200
[Asserts]
jsonpath "$.holder" == ""
jsonpath "$.is_expired" == true

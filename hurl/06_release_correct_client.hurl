POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.job" == "default"

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200
[Asserts]
jsonpath "$.message" == "lock released"
jsonpath "$.job" == "default"

GET http://localhost:8080/lock
HTTP 200
[Asserts]
jsonpath "$.holder" == ""
jsonpath "$.job" == "default"
jsonpath "$.is_expired" == true

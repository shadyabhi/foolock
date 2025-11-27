POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "default"
jsonpath "$.expires_at" exists

GET http://localhost:8080/lock
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "default"
jsonpath "$.expires_at" exists

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200
[Asserts]
jsonpath "$.job" == "default"

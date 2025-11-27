POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "default"

POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "default"
jsonpath "$.message" == "renewed"

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200

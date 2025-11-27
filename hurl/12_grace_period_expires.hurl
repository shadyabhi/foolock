POST http://localhost:8080/lock?client=laptop1&ttl=2s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "default"

GET http://localhost:8080/lock
[Options]
delay: 8000
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "default"
jsonpath "$.is_expired" == true

POST http://localhost:8080/lock?client=laptop2&ttl=5s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop2"
jsonpath "$.job" == "default"

DELETE http://localhost:8080/lock?client=laptop2
HTTP 200

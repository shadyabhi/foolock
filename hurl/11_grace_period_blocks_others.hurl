POST http://localhost:8080/lock?client=laptop1&ttl=2s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"

GET http://localhost:8080/lock
[Options]
delay: 3000
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.is_expired" == true
jsonpath "$.grace_until" exists

POST http://localhost:8080/lock?client=laptop2&ttl=5s
HTTP 409
[Asserts]
jsonpath "$.error" == "grace period active"

POST http://localhost:8080/lock?client=laptop1&ttl=2s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200

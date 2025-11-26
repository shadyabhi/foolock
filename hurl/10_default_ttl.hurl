POST http://localhost:8080/lock?client=laptop1
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.expires_at" exists

DELETE http://localhost:8080/lock?client=laptop1
HTTP 200

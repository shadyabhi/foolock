POST http://localhost:8080/lock?client=test&ttl=invalid
HTTP 400
[Asserts]
jsonpath "$.error" == "invalid ttl format"

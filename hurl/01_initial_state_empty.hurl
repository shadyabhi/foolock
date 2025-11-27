GET http://localhost:8080/lock
HTTP 200
[Asserts]
jsonpath "$.holder" == ""
jsonpath "$.job" == "default"
jsonpath "$.is_expired" == true

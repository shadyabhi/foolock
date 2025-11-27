# Test that different jobs are isolated from each other
# Client1 acquires lock on job1
POST http://localhost:8080/lock?client=laptop1&job=job1&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "job1"

# Client2 can acquire lock on job2 (different job)
POST http://localhost:8080/lock?client=laptop2&job=job2&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop2"
jsonpath "$.job" == "job2"

# Client2 cannot acquire lock on job1 (already held by laptop1)
POST http://localhost:8080/lock?client=laptop2&job=job1&ttl=10s
HTTP 409
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "job1"

# Status for job1 shows laptop1
GET http://localhost:8080/lock?job=job1
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop1"
jsonpath "$.job" == "job1"

# Status for job2 shows laptop2
GET http://localhost:8080/lock?job=job2
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop2"
jsonpath "$.job" == "job2"

# Release job1
DELETE http://localhost:8080/lock?client=laptop1&job=job1
HTTP 200
[Asserts]
jsonpath "$.job" == "job1"

# Now client2 can acquire job1
POST http://localhost:8080/lock?client=laptop2&job=job1&ttl=10s
HTTP 200
[Asserts]
jsonpath "$.holder" == "laptop2"
jsonpath "$.job" == "job1"

# Cleanup
DELETE http://localhost:8080/lock?client=laptop2&job=job1
HTTP 200

DELETE http://localhost:8080/lock?client=laptop2&job=job2
HTTP 200

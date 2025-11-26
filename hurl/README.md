# E2E testing

We're using the tool `hurl` (https://hurl.dev/) to do E2E testing.

## Run

```bash
$ make run
hurl --test --jobs 1 *.hurl
Success 01_initial_state_empty.hurl (1 request(s) in 1 ms)
Success 02_acquire_lock.hurl (3 request(s) in 0 ms)
Success 03_renew_lock.hurl (3 request(s) in 0 ms)
Success 04_conflict_different_client.hurl (3 request(s) in 0 ms)
Success 05_release_wrong_client.hurl (3 request(s) in 0 ms)
Success 06_release_correct_client.hurl (3 request(s) in 0 ms)
Success 07_error_missing_client.hurl (2 request(s) in 0 ms)
Success 08_error_invalid_ttl.hurl (1 request(s) in 0 ms)
Success 09_error_method_not_allowed.hurl (1 request(s) in 0 ms)
Success 10_default_ttl.hurl (2 request(s) in 0 ms)
Success 11_grace_period_blocks_others.hurl (5 request(s) in 3007 ms)
Success 12_grace_period_expires.hurl (4 request(s) in 8006 ms)
--------------------------------------------------------------------------------
Executed files:    12
Executed requests: 31 (2.8/s)
Succeeded files:   12 (100.0%)
Failed files:      0 (0.0%)
Duration:          11023 ms (0h:0m:11s:23ms)
```

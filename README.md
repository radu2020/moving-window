## MovingWindowService

MovingWindowService is a simple Go HTTP server that on each request responds with a
counter of the total number of requests that it has received during the previous 60 seconds
(moving window).

The server should continue to return the correct numbers after restarting it,
by persisting data to a file.

## Description
The service is using an in memory cache to count the requests.

There is a graceful shutdown which also writes the cache to a local file
inside a SQLite DB.

On startup the service looks if there is any recent data inside the DB which
might be worth picking up and if so it then loads this data back into the cache
so that new incoming requests would receive an accurate request count response.

## Installation

```bash
$ go mod tidy
```

## Running the app

```bash
$ go run main.go
```

## Request output example
```bash
$ curl -X Post localhost:3000/count
{"count":9}
$ curl -X Post localhost:3000/count
{"count":3}
```

## Service startup and shutdown output example
```bash
$ go run main.go
2023/08/23 16:44:24 8 entries were recovered from DB and loaded into memory.
2023/08/23 16:44:24 Server available on http://localhost:3000/count

^C2023/08/23 16:44:31 shutting down
2023/08/23 16:44:31 cleaning up: http-server
2023/08/23 16:44:31 cleaning up: database
2023/08/23 16:44:31 http-server was shutdown gracefully
2023/08/23 16:44:31 database was shutdown gracefully

```

## Running the tests

```bash
$ go test -v
```

## Test output
```bash
=== RUN   TestMainProgram
=== RUN   TestMainProgram/should_return_correct_request_amount_for_each_request
2023/08/28 14:11:33 0 entries were recovered from DB and loaded into memory.
2023/08/28 14:11:33 Server available on http://localhost:3000/count
=== RUN   TestMainProgram/should_return_correct_request_amount_after_60_seconds_have_passed
Waiting 61 seconds...
--- PASS: TestMainProgram (61.00s)
--- PASS: TestMainProgram/should_return_correct_request_amount_for_each_request (0.00s)
--- PASS: TestMainProgram/should_return_correct_request_amount_after_60_seconds_have_passed (61.00s)
PASS
ok      MovingWindowRequest     61.248s
```

## Thank you!
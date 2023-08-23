## MovingWindowService

MovingWindowService is a simple Go HTTP server that on each request responds with a
counter of the total number of requests that it has received during the previous 60 seconds
(moving window). The server should continue to return the correct numbers after restarting it,
by persisting data to a file.

## Description
The service is using an in memory cache to count the requests.
There is a graceful shutdown which also write the cache to a local file
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

## Thank you!
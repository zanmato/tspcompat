# tspcompat

Transforms the new format into the legacy app format.

## Build

```sh
$ make linux
```

## Development

```sh
$ go run cmd/proxy/main.go --new-url=http://localhost/test.json --old-url=http://localhost/old.json
$ DB_DSN=postgres://tspcompat:tspcompat@localhost:5450/tspcompat?sslmode=disable\&timezone=Europe/Stockholm go run cmd/api/main.go --load-from=http://localhost/test.json
```

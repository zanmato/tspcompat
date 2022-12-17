# tspcompat

Transforms the new format into the legacy app format.

## Build

```sh
$ make linux
```

## Development

```sh
$ go run cmd/proxy/main.go --new-url=http://localhost/test.json --old-url=http://localhost/old.json
```

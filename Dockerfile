FROM golang:1.19-alpine as builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
ENV USER=tsp
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735RUN 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

WORKDIR $GOPATH/src/tspcompat/
COPY . .

# Fetch dependencies
RUN go mod download

# Build the binaries
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/tspcompatproxy ./cmd/proxy/main.go
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/tspcompatapi ./cmd/api/main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/tspcompatproxy /go/bin/tspcompatproxy
COPY --from=builder /go/bin/tspcompatapi /go/bin/tspcompatapi
USER tsp:tsp
ENTRYPOINT ["/go/bin/tspcompatproxy"]
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

# Build the binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/tspcompat ./cmd/proxy/main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/tspcompat /go/bin/tspcompat
USER tsp:tsp
ENTRYPOINT ["/go/bin/tspcompat"]
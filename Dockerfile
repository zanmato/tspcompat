FROM node:18-alpine as frontend-builder

WORKDIR /app
COPY frontend/package.json .
COPY frontend/yarn.lock .
RUN yarn install

RUN touch /app/.env && echo "VITE_API_URL=" >> /app/.env

COPY frontend/public /app/public
COPY frontend/vite.config.js .
COPY frontend/src /app/src
COPY frontend/index.html /app/index.html
RUN yarn build

FROM golang:1.20-alpine as backend-builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
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

WORKDIR $GOPATH/src/tsp/
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-extldflags "-static" -w -s' -o /app/tsp ./cmd/api/main.go ./cmd/api/migrate.go

FROM scratch
ENV TZ=Europe/Stockholm
COPY --from=backend-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /etc/passwd /etc/passwd
COPY --from=backend-builder /etc/group /etc/group
COPY --from=backend-builder /app/tsp /app/tsp
COPY --from=frontend-builder /app/dist /app/dist
USER tsp:tsp
ENTRYPOINT ["/app/tsp"]
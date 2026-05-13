FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1001 appuser

COPY --from=builder /app/server /app/server

USER appuser
WORKDIR /app

EXPOSE 8080

ENTRYPOINT ["/app/server"]

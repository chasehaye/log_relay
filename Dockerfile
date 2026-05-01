FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server/main.go


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 30000

CMD ["/app/server"]

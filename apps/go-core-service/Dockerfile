FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .docker

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/api .

EXPOSE 8080

CMD ["/app/api"]
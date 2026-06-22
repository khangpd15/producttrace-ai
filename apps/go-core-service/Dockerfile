FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache tzdata ca-certificates

ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/api .

EXPOSE 8080

CMD ["/app/api"]
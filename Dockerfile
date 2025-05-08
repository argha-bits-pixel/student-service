FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
COPY .env /app/.env
RUN go mod tidy
RUN go build -o ./student-service cmd/main.go
 
FROM alpine:latest
WORKDIR /code
COPY --from=builder /app/student-service ./student-service
ENTRYPOINT ["./student-service"]
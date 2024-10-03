FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-mgmt-service ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /user-mgmt-service .
EXPOSE 8080
CMD ["./user-mgmt-service"]
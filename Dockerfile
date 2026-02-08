FROM golang:1.21-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/testbarn ./main.go

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/testbarn /usr/local/bin/testbarn
COPY config.yml ./config.yml
EXPOSE 8080
CMD ["/usr/local/bin/testbarn"]

FROM base AS test
CMD ["go", "test", "-v", "./..."]

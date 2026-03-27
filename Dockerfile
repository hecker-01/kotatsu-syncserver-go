# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api .

# Stage 2: run
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/api .
COPY setup.sql .

EXPOSE 9292
CMD ["./api"]
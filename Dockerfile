# Stage 1: Build
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server main.go

# Stage 2: Run
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/uploads ./uploads
EXPOSE 8000
CMD ["./server"]
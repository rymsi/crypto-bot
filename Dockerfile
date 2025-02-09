# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/ingestor ./cmd/ingestor

# Final minimal runtime container
FROM alpine:3.17
WORKDIR /app
COPY --from=build /bin/ingestor /app/
ENTRYPOINT ["/app/ingestor"]

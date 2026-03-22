# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /matetra-server ./cmd/matetra-server

# Runtime stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /matetra-server /usr/local/bin/matetra-server
EXPOSE 1729
ENV PORT=1729
ENTRYPOINT ["matetra-server", "start"]

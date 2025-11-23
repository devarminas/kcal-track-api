# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.4
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app

# Install certificates for HTTPS module downloads.
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /bin/server /app/server

ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
CMD ["/app/server"]

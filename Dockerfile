# syntax=docker/dockerfile:1

FROM golang:1.22 AS build
WORKDIR /app

# Go module deps (cached)
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

# Build static-ish binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /signalforge ./cmd/app

# Minimal runtime with certs and nonroot user
FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /signalforge /signalforge
ENTRYPOINT ["/signalforge"]
CMD ["-mode", "daily"]

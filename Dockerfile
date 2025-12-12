# syntax=docker/dockerfile:1.6

########## 1) Builder stage ##########
FROM golang:alpine AS build

WORKDIR /app

# Build-time environment
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on

# 1. Cache modules (fast!)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 2. Copy source
COPY . .

# 3. Build optimized binary with cached build folders
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /signalforge ./cmd/app

########## 2) Runtime stage ##########
FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /signalforge /signalforge

ENTRYPOINT ["/signalforge"]
CMD ["-mode", "daily"]

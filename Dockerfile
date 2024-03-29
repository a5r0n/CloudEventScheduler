FROM golang:1.17-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/build

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 go build -o ./out/ ./cmd/...

# Start fresh from a smaller image
FROM alpine:3.9 
RUN apk add ca-certificates --no-cache 

COPY --from=build_base /tmp/build/out/ /app

# This container exposes port 8080 to the outside world
EXPOSE 8080
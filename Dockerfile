FROM golang:1.18-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/invest-app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Unit tests
RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o ./out/invest-app .

# Start fresh from a smaller image
FROM alpine:3.15
ARG TEMP=/tmp/invest-app
COPY --from=build_base $TEMP/out/invest-app /app/invest-app

# Run the binary program produced by `go install`
CMD ["/app/invest-app"]

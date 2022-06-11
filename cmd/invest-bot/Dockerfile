FROM golang:1.18-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/invest-app

# Populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Unit tests
RUN CGO_ENABLED=0 go test -v ./...

# Build the Go app
RUN go build -o ./out/invest-app ./cmd/invest-bot

# Start fresh from a smaller image
FROM alpine:3.15
ARG TEMP=/tmp/invest-app
COPY --from=build_base $TEMP/out/invest-app /app/invest-app
RUN mkdir "/opt/log"
ENV LOG_FILE_PATH="/opt/log/invest-bot.log"
# Volume to mount logs to host (writes in single file - for correct rolling you can use logrotate on host machine)
VOLUME "/opt/log"

RUN apk update && apk add tzdata
ENV LANG ru_RU.UTF-8
ENV LANGUAGE ru_RU:ru
ENV LC_ALL ru_RU.UTF-8
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Run the binary program produced by `go install`
CMD ["/app/invest-app"]

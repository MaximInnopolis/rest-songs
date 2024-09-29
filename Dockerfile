## Build

# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.22 AS build

WORKDIR /cmd

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download
# copy all files
COPY . ./
# build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app -v ./cmd

## Deploy
FROM ubuntu:latest AS final

WORKDIR /

## Install Go
#RUN apk add --no-cache bash

COPY --from=build /bin/app /app

## Install bash for command execution
#RUN apk add --no-cache bash

EXPOSE 8080
EXPOSE 8090

ENTRYPOINT ["/app"]

############################
# STEP 1 build executable binary
# @see https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .

#FIX For  standard_init_linux.go:219: exec user process caused: no such file or directory
ENV CGO_ENABLED=0

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/hello
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN go build -o /go/bin/main
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy CA certificates to prevent x509: certificate signed by unknown authority errors
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Copy our static executable.
COPY --from=builder /go/bin/main /go/bin/main
# Run the hello binary.
ENTRYPOINT ["/go/bin/main"]

FROM golang:alpine as builder


RUN apk update && apk add --no-cache git ca-certificates gcc musl-dev && update-ca-certificates

ENV USER=user
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go mod verify

COPY pkg pkg
env CGO_ENABLED=1
RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/products pkg/examples/products/main.go
RUN chmod +x /go/bin/products
############################
# STEP 2 build a small image
############################
FROM alpine
# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our static executable
COPY --from=builder /go/bin/products /go/bin/products
# Use an unprivileged user.
USER user:user
# Run the products binary.
ENTRYPOINT ["/go/bin/products"]
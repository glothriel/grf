ARG GO_VERSION=1.21.1

# Build stage
FROM golang:${GO_VERSION}-bullseye AS build
 
RUN useradd -u 1000 nonroot -m
USER nonroot:nonroot
WORKDIR /src
COPY ./go.mod ./go.sum ./

RUN go mod download
COPY pkg ./pkg
RUN CGO_ENABLED=1 go build \
    -installsuffix 'static' \
    -o /home/nonroot/app ./pkg/examples/products/main.go

# Dev stage, can be used for development using docker-compose "build.target" config
FROM build as dev
WORKDIR /home/nonroot
ARG VERSION=dev
ENV VERSION=${VERSION}
ENTRYPOINT ["/home/nonroot/app"]
CMD ["start"]

# Prod stage, uses distrolless image
FROM gcr.io/distroless/base AS prod
ARG VERSION=dev
ENV VERSION=${VERSION}
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /home/nonroot/app /home/nonroot/app
LABEL maintainer="https://github.com/glothriel"
ENTRYPOINT ["/home/nonroot/app"]
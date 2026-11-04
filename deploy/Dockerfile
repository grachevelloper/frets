FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY services ./services

ARG SERVICE
RUN --network=none CGO_ENABLED=0 go build -mod=vendor -trimpath -o /frets "./services/${SERVICE}"

FROM alpine:3.23
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 frets \
    && mkdir -p /data/files \
    && chown frets:frets /data/files

COPY --from=build /frets /usr/local/bin/frets
USER frets
ENTRYPOINT ["/usr/local/bin/frets"]

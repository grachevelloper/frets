FROM golang:1.27.3-alpine AS build
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
COPY backend/vendor ./vendor
COPY backend/lib ./lib
COPY backend/services ./services

ARG SERVICE
RUN --network=none CGO_ENABLED=0 go build -mod=vendor -trimpath -o /frets "./services/${SERVICE}/cmd/${SERVICE}"

FROM alpine:3.23
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 frets \
    && mkdir -p /data/files \
    && chown frets:frets /data/files

COPY --from=build /frets /usr/local/bin/frets
USER frets
ENTRYPOINT ["/usr/local/bin/frets"]

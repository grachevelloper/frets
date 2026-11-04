FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY services ./services

ARG SERVICE
RUN CGO_ENABLED=0 go build -trimpath -o /frets "./services/${SERVICE}"

FROM alpine:3.23
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 frets \
    && mkdir -p /data/files \
    && chown frets:frets /data/files

COPY --from=build /frets /usr/local/bin/frets
USER frets
ENTRYPOINT ["/usr/local/bin/frets"]

# builder
FROM golang:1.17.8-alpine3.15 AS builder
RUN apk add build-base
WORKDIR /usr/src/app
COPY go.mod go.sum ./
COPY p2p/ p2p/
COPY util/ util/
COPY cmd/ cmd/
COPY Makefile .
RUN make build

# runner
FROM alpine:3.15.1 AS runner
WORKDIR /usr/bin/app
COPY --from=builder /usr/src/app/kistio-agent .
COPY DOCKER/docker-entrypoint.sh /usr/bin
CMD [ "docker-entrypoint.sh" ]
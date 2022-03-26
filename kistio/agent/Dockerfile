# builder
FROM golang:1.17.8-alpine3.15 AS builder
RUN apk add build-base
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY p2p/ p2p/
COPY server/ server/
COPY util/ util/
COPY cmd/ cmd/
COPY Makefile .
RUN make build

# runner
FROM alpine:3.15.1 AS runner
ARG GRPC_HEALTH_PROBE_VERSION=v0.4.8
RUN apk add dumb-init --no-cache
RUN  wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    chmod +x /bin/grpc_health_probe
WORKDIR /usr/bin/app
COPY --from=builder /usr/src/app/kistio-agent .
COPY DOCKER/docker-entrypoint.sh /usr/bin
CMD [ "docker-entrypoint.sh" ]

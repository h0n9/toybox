# builder
FROM golang:1.18.2-alpine3.16 AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY cli/ cli/
COPY proto/ proto/
COPY store/ store/
COPY lake/ lake/
RUN go build ./cmd/msg-lake

# runner
FROM alpine:3.16.0 AS runner
WORKDIR /usr/bin/app
RUN addgroup --system app && adduser --system --shell /bin/false --ingroup app app
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe
COPY --from=builder /usr/src/app/msg-lake .
RUN chown -R app:app /usr/bin/app
USER app
ENTRYPOINT [ "/usr/bin/app/msg-lake" ]

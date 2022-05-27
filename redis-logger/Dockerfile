# builder
FROM golang:1.18.2-alpine3.16 AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY util/ util/
COPY cmd/ cmd/
RUN go build ./cmd/redis-logger

# runner
FROM alpine:3.16.0 AS runner
WORKDIR /usr/bin/app
RUN addgroup --system app && adduser --system --shell /bin/false --ingroup app app
RUN chown -R app:app /usr/bin/app
COPY --from=builder /usr/src/app/redis-logger .
USER app
ENTRYPOINT [ "/usr/bin/app/redis-logger" ]

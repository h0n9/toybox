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
COPY --from=builder /usr/src/app/msg-lake .
RUN chown -R app:app /usr/bin/app
USER app
ENTRYPOINT [ "/usr/bin/app/msg-lake" ]

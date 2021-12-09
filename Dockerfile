FROM golang:1-alpine as builder

ARG VERSION

RUN go install github.com/korylprince/fileenv@v1.1.0
RUN go install "github.com/korylprince/ls-relay-check@$VERSION"


FROM alpine:3.15

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/fileenv /
COPY --from=builder /go/bin/ls-relay-check /

CMD ["/fileenv", "/ls-relay-check"]

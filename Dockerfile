FROM golang:1.14.4-alpine3.12 AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/image-resize-proxy/
COPY . .

RUN go get -d -v

RUN go build -o /go/bin/image-resize-proxy

#########################
FROM scratch

COPY --from=builder /go/bin/image-resize-proxy /go/bin/image-resize-proxy

ENV PORT 8080
EXPOSE 8080

ENTRYPOINT ["/go/bin/image-resize-proxy"]


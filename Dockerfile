FROM golang:1.14.4-alpine3.12 AS builder

RUN apk add --update --no-cache ca-certificates git

RUN mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/image-resize-proxy
#RUN go build -o /go/bin/image-resize-proxy

#########################
FROM scratch

COPY --from=builder /go/bin/image-resize-proxy /go/bin/image-resize-proxy

ENV PORT 8080
EXPOSE 8080

ENTRYPOINT ["/go/bin/image-resize-proxy"]


FROM golang:1.9 as builder

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...

ENV CGO_ENABLED=0
RUN go build -v -a -o /prometheus-pingdom-exporter

FROM alpine:3.4

COPY --from=builder /prometheus-pingdom-exporter /prometheus-pingdom-exporter

RUN apk update && apk add ca-certificates

EXPOSE 8000

ENTRYPOINT ["/prometheus-pingdom-exporter"]

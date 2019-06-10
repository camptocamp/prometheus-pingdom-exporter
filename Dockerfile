FROM golang:1.12 as builder
MAINTAINER Joseph Salisbury <joseph@giantswarm.io>

WORKDIR /go/src/github.com/giantswarm/prometheus-pingdom-exporter
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go install -a -v \
	-tags netgo \
	-ldflags \
	"-X \"github.com/giantswarm/prometheus-pingdom-exporter/cmd.version=$(cat VERSION)\" \
	 -X \"github.com/giantswarm/prometheus-pingdom-exporter/cmd.goVersion=$(go version)\" \
	 -X \"github.com/giantswarm/prometheus-pingdom-exporter/cmd.gitCommit=$(git rev-parse --short HEAD)\" \
	 -X \"github.com/giantswarm/prometheus-pingdom-exporter/cmd.osArch=$(go env GOOS)/$(go env GOARCH)\" \
	 -w"

FROM gcr.io/distroless/static

COPY --from=builder /go/bin/prometheus-pingdom-exporter /

EXPOSE 8000

ENTRYPOINT ["/prometheus-pingdom-exporter"]

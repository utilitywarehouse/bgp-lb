FROM golang:1.25-alpine AS build
WORKDIR /go/src/github.com/utilitywarehouse/bgp-lb
COPY . /go/src/github.com/utilitywarehouse/bgp-lb
ENV CGO_ENABLED=0
# Skip the tests that need host network
RUN apk --no-cache add git \
      && go get -t ./... \
      && go test --skip PingCheck ./... \
      && go build -o /bgp-lb .

FROM alpine:3.22
COPY --from=build /bgp-lb /bgp-lb

ENTRYPOINT ["/bgp-lb"]

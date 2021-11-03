FROM golang:1.17-alpine AS build
WORKDIR /go/src/github.com/utilitywarehouse/bgp-lb
COPY . /go/src/github.com/utilitywarehouse/bgp-lb
 ENV CGO_ENABLED 0
RUN apk --no-cache add git &&\
  go get -t ./... &&\
  go test ./... &&\
  go build -o /bgp-lb .

FROM alpine:3.14
COPY --from=build /bgp-lb /bgp-lb

ENTRYPOINT ["/bgp-lb"]

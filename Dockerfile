FROM golang:alpine as golang
WORKDIR /go/src/github.com/cloudfoundry/bbl-state-resource/
COPY . .

ENV GOBIN=/go/bin/
RUN go install cmd/check/check.go
RUN go install cmd/out/out.go
RUN go install cmd/in/in.go

FROM alpine
MAINTAINER https://github.com/cloudfoundry/bbl-state-resource

RUN apk add --no-cache ca-certificates

COPY --from=golang /go/bin/check /opt/resource/check
COPY --from=golang /go/bin/in /opt/resource/in
COPY --from=golang /go/bin/out /opt/resource/out

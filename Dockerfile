FROM cloudfoundry/cf-deployment-concourse-task
WORKDIR /go/src/github.com/cloudfoundry/bbl-state-resource/
COPY . .

ENV CGO_ENABLED 0
ENV GOBIN /opt/resource/
RUN go install cmd/check/check.go
RUN go install cmd/out/out.go
RUN go install cmd/in/in.go

LABEL MAINTAINER=https://github.com/cloudfoundry/bbl-state-resource

ARG bbl_version=8.4.89
RUN wget https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${bbl_version}/bbl-v${bbl_version}_linux_x86-64 -O /usr/local/bin/bbl \
    && chmod +x /usr/local/bin/bbl

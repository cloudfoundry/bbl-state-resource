FROM golang:alpine as golang
WORKDIR /go/src/github.com/cloudfoundry/bbl-state-resource/
COPY . .

ENV CGO_ENABLED 0
ENV GOBIN /go/bin/
RUN go install cmd/check/check.go
RUN go install cmd/out/out.go
RUN go install cmd/in/in.go

MAINTAINER https://github.com/cloudfoundry/bbl-state-resource

# Install bosh_cli v2
RUN curl -L "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-$(curl -s https://api.github.com/repos/cloudfoundry/bosh-cli/releases | jq -r .[0].name | tr -d "v")-linux-amd64" -o "/usr/local/bin/bosh" && \
  chmod +x "/usr/local/bin/bosh"

# install (or reinstall?) bosh-cli dependencies.
# we'd love for this to work in a small alpine container, but the cli
# and cpis seem to need ubuntu packages
RUN apt-get update && apt-get -y install software-properties-common && \
    add-apt-repository ppa:brightbox/ruby-ng -y && \
    apt-get update && \
    apt-get -y install ruby2.3 ruby2.3-dev \
                       build-essential \
                       libreadline6 libreadline6-dev \
                       libsqlite3-dev libssl-dev \
                       libxml2-dev libxslt-dev \
                       libyaml-dev openssl \
                       sqlite unzip wget curl zlib1g-dev zlibc && \
    apt-get remove -y --purge software-properties-common

ARG bbl_version=6.2.1
RUN wget https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${bbl_version}/bbl-v${bbl_version}_linux_x86-64 -O /usr/local/bin/bbl \
    && chmod +x /usr/local/bin/bbl

COPY --from=golang /go/bin/check /opt/resource/check
COPY --from=golang /go/bin/in /opt/resource/in
COPY --from=golang /go/bin/out /opt/resource/out

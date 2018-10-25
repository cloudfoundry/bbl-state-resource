FROM relintdockerhubpushbot/cf-deployment-concourse-tasks
WORKDIR /go/src/github.com/cloudfoundry/bbl-state-resource/
COPY . .

ENV CGO_ENABLED 0
ENV GOBIN /go/bin/
RUN go install cmd/check/check.go
RUN go install cmd/out/out.go
RUN go install cmd/in/in.go

MAINTAINER https://github.com/cloudfoundry/bbl-state-resource

ARG bbl_version=6.2.1
RUN wget https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${bbl_version}/bbl-v${bbl_version}_linux_x86-64 -O /usr/local/bin/bbl \
    && chmod +x /usr/local/bin/bbl

COPY --from=golang /go/bin/check /opt/resource/check
COPY --from=golang /go/bin/in /opt/resource/in
COPY --from=golang /go/bin/out /opt/resource/out

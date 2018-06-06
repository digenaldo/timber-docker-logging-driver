# The Mozilla CA Certificate Store is being redistributed within this
# container under the Mozilla Public License 2.0 via the Alpine Linux
# ca-certificates package.
# https://www.mozilla.org/en-US/about/governance/policies/security-group/certs/
# https://www.mozilla.org/media/MPL/2.0/index.815ca599c9df.txt
# https://pkgs.alpinelinux.org/package/edge/main/x86_64/ca-certificates

FROM golang:1.10.2-alpine3.7 as builder

RUN apk --no-cache add curl git ca-certificates
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/timberio/timber-docker-logging-driver/
COPY . .

RUN dep ensure
RUN go build -o /usr/bin/timber-docker-logging-driver .


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt \
     /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/bin/timber-docker-logging-driver \
     /usr/bin/timber-docker-logging-driver

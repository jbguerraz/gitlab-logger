FROM golang:alpine as builder
COPY main.go /tmp/gitlab-logger/main.go
ENV CGO_ENABLED=0
RUN apk add --no-cache git \
    && cd /tmp/gitlab-logger \
    && export GOPATH=/tmp/gitlab-logger \
    && export GOBIN=/usr/bin \
    && git config --global http.https://gopkg.in.followRedirects true \
    && go get \
    && go build .

FROM scratch
COPY --from=builder /tmp/gitlab-logger/gitlab-logger /
ENTRYPOINT ["/gitlab-logger"]
CMD ["--poll=false", "--json=false", "--minlevel=0", "--exclude=sasl|config|lock|@|gzip|tgz|gz", "/var/log/gitlab"]

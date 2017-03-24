FROM golang:1.8

ENV GOPATH /go
ENV GLIDE_VERSION 0.12.3
ENV APP_DIR /go/src/github.com/djmaze/shepherd

RUN curl -sL https://github.com/Masterminds/glide/releases/download/v${GLIDE_VERSION}/glide-v${GLIDE_VERSION}-linux-amd64.tar.gz \
  | tar xzf - --strip-components=1 -C /usr/local/bin linux-amd64/glide

WORKDIR ${APP_DIR}
COPY glide.yaml glide.lock ${APP_DIR}/
RUN glide install

VOLUME ${APP_DIR}/vendor

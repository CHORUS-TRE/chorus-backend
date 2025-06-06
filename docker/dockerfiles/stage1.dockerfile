FROM harbor.build.chorus-tre.ch/chorus/backend-ubuntu:latest

USER root

ARG GOLANG_VERSION=1.24.3
RUN curl -LO https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"

ENV GOCACHE="/chorus/.cache/go-build"
ENV GOMODCACHE="/chorus/.cache/go-mod"

WORKDIR /chorus
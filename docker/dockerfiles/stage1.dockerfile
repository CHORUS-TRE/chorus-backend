FROM registry.dip-dev.thehip.app/chorus-ubuntu:latest

USER root

ARG GOLANG_VERSION=1.22.5
RUN curl -LO https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"

ENV GOCACHE="/chorus/.cache/go-build"
ENV GOMODCACHE="/chorus/.cache/go-mod"

WORKDIR /chorus
FROM registry.dip-dev.thehip.app/chorus-stage1:latest AS build

USER chorus

COPY --chown=chorus:chorus . /chorus

RUN cd /chorus/cmd/chorus && \
    go build -trimpath -ldflags "$LD_FLAGS" -o ../../bin/chorus

FROM ubuntu:latest

COPY --from=build /chorus/bin/chorus /chorus/bin/chorus

CMD ["/chorus/bin/chorus", "start", "--config", "/chorus/conf/config.yaml""]
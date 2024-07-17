FROM registry.dip-dev.thehip.app/chorus-stage1:latest AS build

# USER chorus
# COPY --chown=chorus:chorus . /chorus

COPY . /chorus

ENV GOCACHE="/chorus/.cache/go-build"
ENV GOMODCACHE="/chorus/.cache/go-mod"

RUN --mount=type=cache,target="/chorus/.cache/go-build" --mount=type=cache,target="/chorus/.cache/go-mod" cd /chorus/cmd/chorus && \
    go build -trimpath -ldflags "$LD_FLAGS" -o /chorus/bin/chorus

FROM ubuntu:latest

COPY --from=build /chorus/bin/chorus /chorus/bin/chorus

CMD ["/chorus/bin/chorus", "start", "--config", "/chorus/conf/config.yaml"]
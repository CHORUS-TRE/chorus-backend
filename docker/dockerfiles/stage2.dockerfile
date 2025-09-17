FROM harbor.build.chorus-tre.ch/chorus/backend-stage1:latest AS build

# USER chorus
# COPY --chown=chorus:chorus . /chorus

COPY . /chorus

ENV GOCACHE="/chorus/.cache/go-build"
ENV GOMODCACHE="/chorus/.cache/go-mod"

RUN --mount=type=secret,id=GIT_USERNAME \
    --mount=type=secret,id=GIT_PASSWORD \
    u="$(cat /run/secrets/GIT_USERNAME)"; \
    p="$(cat /run/secrets/GIT_PASSWORD)"; \
    [ -n "$u" ] && [ -n "$p" ] && \
    GOPRIVATE=github.com/CHORUS-TRE/* \
    GIT_CONFIG_COUNT=1 \
    GIT_CONFIG_KEY_0=url."https://${u}:${p}@github.com/".insteadof \
    GIT_CONFIG_VALUE_0=https://github.com/ \
    go mod download

RUN --mount=type=cache,target="/chorus/.cache/go-build" --mount=type=cache,target="/chorus/.cache/go-mod" cd /chorus/cmd/chorus && \
    go build -trimpath -ldflags "$LD_FLAGS" -o /chorus/bin/chorus

FROM ubuntu:latest

RUN apt update && apt install -y ca-certificates

COPY --from=build /chorus/bin/chorus /chorus/bin/chorus

CMD ["/chorus/bin/chorus", "start", "--config", "/chorus/conf/config.yaml"]
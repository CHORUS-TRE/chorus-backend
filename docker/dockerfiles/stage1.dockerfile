FROM registry.dip-dev.thehip.app/ds-ubuntu:latest

USER root

RUN apt update && apt install libpq-dev -y --no-install-recommends

USER ds

COPY ./requirements.txt /template_backend/requirements.txt
COPY ./docker/secret_exec.sh /template_backend/docker/secret_exec.sh
RUN --mount=type=secret,id=PYPI_USERNAME,uid=1000 --mount=type=secret,id=PYPI_PASSWORD,uid=1000 \
    /template_backend/docker/secret_exec.sh pip install -r /template_backend/requirements.txt

WORKDIR /template_backend
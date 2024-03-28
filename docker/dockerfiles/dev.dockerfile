FROM registry.dip-dev.thehip.app/ds-ubuntu:latest

USER root

RUN apt-get update && \
    apt-get install -y python3 python3-pip unzip default-jre libpq-dev

COPY requirements.txt  /ds/ds-deployer/requirements.txt
RUN pip3 install -r requirements.txt
RUN rm /ds/ds-deployer/requirements.txt

USER ds

ENV SHELL /bin/bash
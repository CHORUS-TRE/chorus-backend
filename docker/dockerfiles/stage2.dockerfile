FROM registry.dip-dev.thehip.app/ds-cicd-template-backend-stage1:latest

USER ds

COPY --chown=ds:ds . /template_backend
USER root
RUN ls -lah /template_backend && rm -rf /template_backend/.git
USER ds

RUN --mount=type=secret,id=PYPI_USERNAME,uid=1000 --mount=type=secret,id=PYPI_PASSWORD,uid=1000 \
    /template_backend/docker/secret_exec.sh pip install -r /template_backend/requirements.txt

CMD ["python3", "-m", "src.cmd.start"]
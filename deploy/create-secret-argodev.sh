kubectl create secret generic backend-secrets \
    --from-file=secrets.yaml=../configs/argodev/secrets.dec.yaml \
    -n backend
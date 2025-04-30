kubectl create secret generic backend-secrets \
    --from-file=secrets.yaml=../configs/argoqa/secrets.dec.yaml \
    -n backend
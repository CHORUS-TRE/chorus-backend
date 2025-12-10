#!/bin/bash

set -e
HELM_REL_NAME="chorus-jenkins-backend"

helm version
helm dependency update ./backend
helm template $HELM_REL_NAME ./backend \
  --namespace "$env" \
  --values ./backend/values.yaml \
  --values ../configs/$env/values.yaml \
  --set-string "image.tag=${IMAGE_TAG}"

echo ""
echo "deploying..."
helm upgrade --install $HELM_REL_NAME ./backend \
  --namespace "$env" \
  --values ./backend/values.yaml \
  --values ../configs/$env/values.yaml \
  --set-string "image.tag=${IMAGE_TAG}"

echo "done"

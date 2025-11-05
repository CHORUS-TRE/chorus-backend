#!/bin/bash

set -e

helm version
helm dependency update ./backend
helm template --namespace "$env" --values ../configs/$env/values.yaml --set-string "image.tag=${IMAGE_TAG}" ./backend

echo "\ndeploying..."
helm upgrade --install --create-namespace --namespace "$env" --values ./backend/files/values.yaml --set-string "version=$DEPLOY_VERSION" --set-string "image.tag=${IMAGE_TAG}" "${RELEASE_NAME}" ./backend
echo "done"
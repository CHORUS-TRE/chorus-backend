#!/bin/bash

set -e

mkdir -p ./backend/files
rm -rf ./backend/files/*
cp -r ../configs/$env/* ./backend/files/

helm version
CONFIG_AES_PASSPHRASE_VAR_NAME="CONFIG_AES_PASSPHRASE_$env"
helm template --namespace "$env" --values ./backend/files/values.yaml --set-string "aesPassphrase=${!CONFIG_AES_PASSPHRASE_VAR_NAME}" --set-string "image=registry.dip-dev.thehip.app/chorus-cicd-chorus:${IMAGE_TAG}" ./backend

echo ""
echo "deploying..."
helm upgrade --install --create-namespace --namespace "$env" --values ./backend/files/values.yaml --set-string "aesPassphrase=${!CONFIG_AES_PASSPHRASE_VAR_NAME}" --set-string "version=$IMAGE_TAG" --set-string "image=registry.dip-dev.thehip.app/chorus-cicd-chorus:${IMAGE_TAG}" "${RELEASE_NAME}" ./backend
echo "done"
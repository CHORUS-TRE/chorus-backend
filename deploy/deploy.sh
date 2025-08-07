#!/bin/bash

set -e

mkdir -p ./backend/files
rm -rf ./backend/files/*
cp -r ../configs/$env/* ./backend/files/

helm version
CONFIG_AES_PASSPHRASE_VAR_NAME="CONFIG_AES_PASSPHRASE_$env"
helm template --namespace "$env" --values ./backend/files/values.yaml --set-string "aesPassphrase=${!CONFIG_AES_PASSPHRASE_VAR_NAME}" --set-string "image=harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG}" ./backend

echo "CONFIG_AES_PASSPHRASE"
echo $CONFIG_AES_PASSPHRASE_VAR_NAME
echo ${!CONFIG_AES_PASSPHRASE_VAR_NAME}

echo ""
echo "deploying..."
helm upgrade --install --create-namespace --namespace "$env" --values ./backend/files/values.yaml --set-string "aesPassphrase=${!CONFIG_AES_PASSPHRASE_VAR_NAME}" --set-string "version=$DEPLOY_VERSION" --set-string "image=harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG}" "${RELEASE_NAME}" ./backend
echo "done"
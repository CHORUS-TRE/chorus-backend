#!/bin/bash

set -e

helm version
helm dependency update ./backend
helm template chorus-jenkins-backend ./backend --namespace "$env" --values ./backend/values.yaml --values ../configs/$env/values.yaml --set-string "image.tag=${IMAGE_TAG}" > chorus-jenkins-backend.yaml

echo ""
echo "deploying..."
kubectl apply -f chorus-jenkins-backend.yaml --namespace "$env"
echo "done"
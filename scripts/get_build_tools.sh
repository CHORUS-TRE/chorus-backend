#!/bin/bash

# Main procedure.
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"


mkdir -p scripts/tools/linux/bin
mkdir -p scripts/tools/darwin/bin


echo
echo "--- getting protoc ---"
PB_REL="https://github.com/protocolbuffers/protobuf/releases"
curl -LO $PB_REL/download/v27.2/protoc-27.2-linux-x86_64.zip
unzip -o protoc-27.2-linux-x86_64.zip -d protoc-27.2-linux-x86_64
cp protoc-27.2-linux-x86_64/bin/protoc scripts/tools/linux/bin
rm -rf protoc-27.2-linux-x86_64*
curl -LO $PB_REL/download/v27.2/protoc-27.2-osx-universal_binary.zip 
unzip -o protoc-27.2-osx-universal_binary.zip -d protoc-27.2-osx-universal_binary
cp protoc-27.2-osx-universal_binary/bin/protoc scripts/tools/darwin/bin
rm -rf protoc-27.2-osx-universal_binary*


echo
echo "--- getting protoc-gen-openapiv2 ---"
curl -L -o scripts/tools/linux/bin/protoc-gen-openapiv2 https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.20.0/protoc-gen-openapiv2-v2.20.0-linux-x86_64
curl -L -o scripts/tools/darwin/bin/protoc-gen-openapiv2 https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.20.0/protoc-gen-openapiv2-v2.20.0-darwin-arm64
chmod u+x scripts/tools/linux/bin/protoc-gen-openapiv2


# echo
# echo "--- getting swagger-codegen ---"
# curl -L -o scripts/tools/swagger-codegen-cli.jar https://repo1.maven.org/maven2/io/swagger/codegen/v3/swagger-codegen-cli/3.0.43/swagger-codegen-cli-3.0.43.jar


echo
echo "--- getting openapi generator ---"
curl -L -o scripts/tools/openapi-generator-cli.jar https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/7.0.0/openapi-generator-cli-7.0.0.jar


echo
echo "==> Installing protoc-gen-go:"
git clone -c advice.detachedHead=false --branch v1.5.4 https://github.com/golang/protobuf.git
cd protobuf/protoc-gen-go
for GOOS in darwin linux; do
  export GOOS=$GOOS
  go build -o $DIR/scripts/tools/$GOOS/bin/protoc-gen-go
done
cd -
rm -rf protobuf

echo
echo "==> Installing protoc-gen-grpc-gateway:"
git clone -c advice.detachedHead=false --branch v2.20.0 https://github.com/grpc-ecosystem/grpc-gateway.git
cd grpc-gateway/protoc-gen-grpc-gateway
for GOOS in darwin linux; do
  export GOOS=$GOOS
  go build -o $DIR/scripts/tools/$GOOS/bin/protoc-gen-grpc-gateway
done
cd -
rm -rf grpc-gateway

echo
echo "==> Installing swagger ($OS):"
curl -o $DIR/scripts/tools/linux/bin/goswagger -L'#' "https://github.com/go-swagger/go-swagger/releases/download/v0.31.0/swagger_linux_amd64"
curl -o $DIR/scripts/tools/darwin/bin/goswagger -L'#' "https://github.com/go-swagger/go-swagger/releases/download/v0.31.0/swagger_darwin_arm64"
chmod +x $DIR/scripts/tools/linux/bin/goswagger
chmod +x $DIR/scripts/tools/darwin/bin/goswagger

for GOOS in darwin linux; do
  chmod -R u+w $DIR/scripts/tools/$GOOS/bin
done

echo
echo "==> Installing golangci-lint:"
v="1.59.1"
# darwin
curl -o golangci-lint-$v-darwin-arm64.tar.gz -L https://github.com/golangci/golangci-lint/releases/download/v$v/golangci-lint-$v-darwin-arm64.tar.gz
tar xzvf golangci-lint-$v-darwin-arm64.tar.gz
mv golangci-lint-$v-darwin-arm64/golangci-lint  $DIR/scripts/tools/darwin/bin/golangci-lint
chmod +x $DIR/scripts/tools/darwin/bin/golangci-lint
#linux
curl -o golangci-lint-$v-linux-amd64.tar.gz -L https://github.com/golangci/golangci-lint/releases/download/v$v/golangci-lint-$v-linux-amd64.tar.gz
tar xzvf golangci-lint-$v-linux-amd64.tar.gz
mv golangci-lint-$v-linux-amd64/golangci-lint  $DIR/scripts/tools/linux/bin/golangci-lint
chmod +x $DIR/scripts/tools/linux/bin/golangci-lint
rm -rf golangci-lint-$v-linux-amd64 golangci-lint-$v-linux-amd64.tar.gz golangci-lint-$v-darwin-arm64 golangci-lint-$v-darwin-arm64.tar.gz

echo 
echo "==> getting swagger ui"
mkdir -p $DIR/api/openapiv2/ui
curl -LO https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.17.14.zip
unzip -o v5.17.14.zip -d v5.17.14
cp -r v5.17.14/swagger-ui-5.17.14/dist/* $DIR/api/openapiv2/ui/
rm -rf v5.17.14*
sed -i -e 's|https://petstore.swagger.io/v2/swagger.json|/openapi/openapiv2/v1-tags/apis.swagger.yaml|g' $DIR/api/openapiv2/ui/swagger-initializer.js


echo 
echo "==> installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
mv kubectl $DIR/scripts/tools/linux/bin
chmod +x $DIR/scripts/tools/linux/bin/kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl"
mv kubectl $DIR/scripts/tools/darwin/bin
chmod +x $DIR/scripts/tools/darwin/bin/kubectl


echo
echo "==> installing kind"
curl -o $DIR/scripts/tools/linux/bin/kind -L'#' "https://github.com/kubernetes-sigs/kind/releases/download/v0.23.0/kind-linux-amd64"
chmod +x $DIR/scripts/tools/linux/bin/kind
curl -o $DIR/scripts/tools/darwin/bin/kind -L'#' "https://github.com/kubernetes-sigs/kind/releases/download/v0.23.0/kind-darwin-arm64"
chmod +x $DIR/scripts/tools/darwin/bin/kind



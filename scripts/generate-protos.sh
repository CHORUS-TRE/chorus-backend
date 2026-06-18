#!/bin/bash

# Main procedure.
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"


OS=darwin
if [[ $(uname -s) == Linux ]]
then
    OS=linux
fi

PATH="$PATH:$PWD/scripts/tools/$OS/bin"

function generate_api_files() {
    # Protobuf and openapiv2 instantiations.
    echo
    echo "==> Handling proto files:"

    mkdir -p api/openapiv2/v1-tags

    rm -rf api/openapiv2/v1-tags/*-service

    for file in api/proto/v1/*.proto; do
        if [[ -f $file ]]; then
            base=$(basename "$file")

            # Go types are needed for every proto (messages used across the codebase).
            echo "---> generating grpc files for $base ..."
            protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --go_out=plugins=grpc:internal/api/v1/chorus "$base"

            # Only *-service.proto declare endpoints: generate the gateway and the
            # per-service OpenAPI (consumed by goswagger in generate_client) for those
            # only. /doc and the frontend use only the merged apis.swagger.yaml below.
            case "$base" in
                *-service.proto)
                    echo "---> generating grpc gateway files for $base ..."
                    protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --grpc-gateway_out=logtostderr=true:internal/api/v1/chorus "$base"

                    filename="${base%.*}"
                    mkdir -p api/openapiv2/v1-tags/$filename
                    protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --openapiv2_out=logtostderr=true,allow_merge=true,output_format=yaml,disable_default_errors=true,merge_file_name=apis:api/openapiv2/v1-tags/$filename "$base"
                    ;;
            esac
        fi
    done

    echo "---> generating merged openapiv2 API definition file 'apis.swagger.yaml' ..."
    protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --openapiv2_out=logtostderr=true,allow_merge=true,output_format=yaml,disable_default_errors=true,merge_file_name=apis:api/openapiv2/v1-tags api/proto/v1/*.proto
}

# function generate_server() {
#     # Protobuf and openapiv2 instantiations.
#     echo
#     echo "==> Handling openapi file:"

#     echo "---> generating flask server ..."
#     java -jar ./scripts/tools/openapi-generator-cli.jar generate \
#        -i api/openapiv2/v1-tags/apis.swagger.yaml \
#        -g python-flask \
#        -o src/internal/api/server_template_tmp \
#     #    -t src/internal/api/generator_template/python-flask \
#        --additional-properties=packageName=server_template

#     rm -rf src/internal/api/server_template
#     mv src/internal/api/server_template_tmp/server_template src/internal/api/server_template
#     rm -r src/internal/api/server_template_tmp
# }

function generate_client() {
    basepath=$(pwd)

    rm -rf tests/helpers/generated/client

    for folder in api/openapiv2/v1-tags/*-service; do
        if [[ -d $folder && -f "$folder/apis.swagger.yaml" ]]; then
            service=$(echo ${folder##*/} | sed 's/-service$//')
            echo "generating openapi client for $service" 

            mkdir -p tests/helpers/generated/client/$service
            cd tests/helpers/generated/client/$service
            goswagger generate client -f $basepath/api/openapiv2/v1-tags/$service-service/apis.swagger.yaml 
            cd -
        fi
    done
}

generate_api_files
generate_client

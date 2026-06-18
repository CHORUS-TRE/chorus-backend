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

echo $PATH

function generate_api_files() {
    # Protobuf and openapiv2 instantiations.
    echo
    echo "==> Handling proto files:"

    mkdir -p api/openapiv2/v1-tags

    for file in api/proto/v1/*.proto; do
        if [[ -f $file ]]; then
            echo "---> generating grpc files for $(basename $file) ..."
            protoc --proto_path=api/proto/v1/ --proto_path=api/proto/third_party --go_out=plugins=grpc:internal/api/v1/chorus `basename $file`

            echo "---> generating grpc gateway files ..."
            protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --grpc-gateway_out=logtostderr=true:internal/api/v1/chorus `basename $file`

            # Per-service OpenAPI files feed goswagger (generate_client) to build the
            # Go acceptance-test HTTP clients. Only *-service.proto declare endpoints;
            # message-only protos produce empty swagger consumed by nothing, so skip
            # them here. The frontend and /doc use only the merged apis.swagger.yaml.
            case "$(basename $file)" in
                *-service.proto)
                    filename=$(basename -- "$file")
                    filename="${filename%.*}"
                    mkdir -p api/openapiv2/v1-tags/$filename
                    protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --openapiv2_out=logtostderr=true,allow_merge=true,output_format=yaml,disable_default_errors=true,merge_file_name=apis:api/openapiv2/v1-tags/$filename $file
                    ;;
            esac
        fi
    done

    echo "---> generating merged openapiv2 API definition file 'apis.openapiv2.json' ..."
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
        if [[ -d $folder && $folder/apis.swagger.yaml ]]; then
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

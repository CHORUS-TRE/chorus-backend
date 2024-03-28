#!/bin/bash

# Main procedure.
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"

PATH="$PATH:$PWD/scripts/tools/linux/bin"

echo $PATH

function generate_api_files() {
    # Protobuf and openapiv2 instantiations.
    echo
    echo "==> Handling proto files:"

    mkdir -p api/openapiv2/v1-tags

    for file in api/proto/v1/*.proto; do
        if [[ -f $file ]]; then
            echo "---> generating grpc files..."
            echo $(basename $file)
            echo $file
            #python3 -m grpc_tools.protoc --proto_path=. --proto_path=api/proto/third_party --python_out=internal --grpc_python_out=internal $file
            # protoc --proto_path=api/proto/v1/ --proto_path=api/proto/third_party --go_out=plugins=grpc:internal/api/v1/templatebackend $(basename $file)


            # echo "---> generating grpc gateway files..."
            # protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --grpc-gateway_out=logtostderr=true:internal/api/v1/templatebackend $(basename $file)

            #echo "---> generating grpc validation files..."
            #protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --govalidators_out=internal/api/v1/templatebackend  `basename $file`

            # echo "---> generating openapiv2 files..."
            # protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --openapiv2_out=disable_default_errors=true,simple_operation_ids=true,logtostderr=true:api/openapiv2/v1-tags $(basename $file)

            filename=$(basename -- "$file")
            filename="${filename%.*}"
            mkdir -p api/openapiv2/v1-tags/$filename
            protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --openapiv2_out=logtostderr=true,allow_merge=true,output_format=yaml,merge_file_name=apis:api/openapiv2/v1-tags/$filename $file
            openapi typegen api/openapiv2/v1-tags/$filename/apis.swagger.yaml > api/openapiv2/v1-tags/$filename/type.d.ts
        fi
    done

    echo "---> generating merged openapiv2 API definition file 'apis.openapiv2.json' ..."
    protoc --proto_path=api/proto/v1 --proto_path=api/proto/third_party --openapiv2_out=logtostderr=true,allow_merge=true,output_format=yaml,merge_file_name=apis:api/openapiv2/v1-tags api/proto/v1/*.proto
}

function generate_types() {
    # Protobuf and openapiv2 instantiations.
    echo
    echo "==> Handling openapi file:"

    npm run openapi
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
    # Protobuf and openapiv2 instantiations.
    echo
    echo "==> Handling openapi file:"

    echo "---> generating flask server ..."
    java -jar ./scripts/tools/openapi-generator-cli.jar generate \
       -i api/openapiv2/v1-tags/apis.swagger.yaml \
       -g typescript-axios \
       -o src/tests/client 

    # rm -rf src/internal/api/server_template
    # mv src/internal/api/server_template_tmp/server_template src/internal/api/server_template
    # rm -r src/internal/api/server_template_tmp
}

generate_api_files
generate_types
generate_client

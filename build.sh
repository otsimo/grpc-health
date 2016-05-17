#!/usr/bin/env bash

NUMBER_OF_COMMIT=$(git rev-list HEAD --count)
SHA_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date +%Y%m%d-%H%M%S)
APP_VERSION=${APP_VERSION:-$(< ./VERSION)}.${NUMBER_OF_COMMIT}
BUILD_VERSION=${APP_VERSION}-${SHA_COMMIT}-${BUILD_DATE}

BUILD_TARGET="$1"

LDFAGS="-w -X main.Version=$BUILD_VERSION"
app_name="grpc-health"

BUILD_OS=$(go env GOOS)
BUILD_ARCH=$(go env GOARCH)

if [ "$BUILD_TARGET" = "docker" ]; then
    BUILD_OS="linux"
    BUILD_ARCH="amd64"
fi
FILE_NAME="${app_name}-${BUILD_OS}-${BUILD_ARCH}"

echo "build $app_name for $BUILD_OS-$BUILD_ARCH"
env CGO_ENABLED=0 GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} go build -ldflags "$LDFAGS" -o bin/${FILE_NAME} main.go

if [ "$?" -ne "0" ];then
    echo failed to build ${app_name}
    exit 2
fi

TAG_SUFFIX="-$2"
if [ "$2" = "prod" ]; then
    TAG_SUFFIX=""
fi
IMG_ORG="$3"

if [ "$BUILD_TARGET" = "docker" ]; then
    cp Dockerfile bin/Dockerfile
    cd bin

    DOCKER_IMAGE_NAME="${IMG_ORG}/${app_name}"
    TAG_NAME=${DOCKER_IMAGE_NAME}:${APP_VERSION}${TAG_SUFFIX}
    docker build --rm -t ${TAG_NAME} .
    echo docker build ${TAG_NAME}
fi
#!/bin/bash
set -euo pipefail

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
BUILD_DIR="$SCRIPTPATH/../../build"

CLI_BINARY_NAME="ec2-instance-qualifier"
AGENT_BINARY_NAME="agent"
APP_BINARY_NAME="ec2-instance-qualifier-app"
LICENSE_TEST_TAG="aeiq-license-test"

make -s -f $SCRIPTPATH/../../Makefile compile app
docker build --build-arg=GOPROXY=direct -t $LICENSE_TEST_TAG $SCRIPTPATH/
docker run -it -e GITHUB_TOKEN --rm -v $SCRIPTPATH/:/test -v $BUILD_DIR/:/build $LICENSE_TEST_TAG golicense /test/license-config.hcl /build/$CLI_BINARY_NAME /build/$AGENT_BINARY_NAME /build/$APP_BINARY_NAME
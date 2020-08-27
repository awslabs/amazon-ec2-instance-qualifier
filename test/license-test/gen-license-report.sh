#!/bin/bash
set -euo pipefail

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
BUILD_DIR="$SCRIPTPATH/../../build"
mkdir -p $BUILD_DIR
GOBIN=$(go env GOPATH | sed 's+:+/bin+g')/bin
export PATH="$PATH:$GOBIN"

go get github.com/mitchellh/golicense
make -s -f $SCRIPTPATH/../../Makefile compile app
golicense -out-xlsx=$BUILD_DIR/report.xlsx $SCRIPTPATH/license-config.hcl $BUILD_DIR/ec2-instance-qualifier $BUILD_DIR/agent $BUILD_DIR/ec2-instance-qualifier-app
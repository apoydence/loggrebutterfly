#!/bin/bash

dir_resolve()
{
    cd "$1" 2>/dev/null || return $?  # cd to desired directory; if fail, quell any error messages but return exit status
    echo "`pwd -P`" # output full, link-resolved path
}

set -e

loggregator=github.com/poy/loggrebutterfly/api/loggregator/v2
target=`dirname $0`
target=`dir_resolve $target`
cd $target

go get github.com/golang/protobuf/{proto,protoc-gen-go}

tmp_dir=$(mktemp -d)
mkdir -p $tmp_dir/$loggregator

cp $GOPATH/src/github.com/cloudfoundry/loggregator-api/v2/*proto $tmp_dir/$loggregator

protoc *.proto --go_out=plugins=grpc:. --proto_path=$tmp_dir --proto_path=.

rm -r $tmp_dir

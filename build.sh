#!/bin/bash
set -e

outfile=$1
verflag=$2
workdir=$3

pushd $workdir
/usr/bin/redis-server &
go test ./... -v
go build -o $outfile -ldflags "-X main.version=$verflag -extldflags '-static'"
# now set build dir to 777 so jenkins user can delete stuff in the volume
chmod -R 777 ./build
ls -al ./build/


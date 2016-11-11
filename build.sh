#!/usr/bin/env bash

RELEASE=0.2.0
dist=dist
bin=oauth-proxy

function build {
    GOOS=$1 GOARCH=$2 RUNMODE=production go build -o $bin
    package=$bin-$RELEASE-$1-$2.tar.gz
    tar cvzf $package $bin
    mv $package $dist
    rm $bin
}

goru dist
mkdir -p $dist
build darwin amd64
build linux amd64

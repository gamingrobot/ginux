#!/bin/bash
echo "I am the build script that is ran on ginux-build."
gvm install go1.2
declare -x PATH="/root/.gvm/pkgsets/go1.2/global/bin:/root/.gvm/gos/go1.2/bin:/root/.gvm/pkgsets/go1.2/global/overlay/bin:/root/.gvm/bin:/root/.gvm/bin:/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin"
declare -x GOPATH="/root/.gvm/pkgsets/go1.2/global"
declare -x GOROOT="/root/.gvm/gos/go1.2"
cd server/game_server
go get
go build
cd ../vzcontrol
go get
go build
cd ../../client
npm install
grunt

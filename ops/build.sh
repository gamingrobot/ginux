#!/bin/bash
echo "I am the build script that is ran on ginux-build."
cd server/game_server
go get
go build
cd ../vzcontrol
go get
go build
cd ../../client
npm install
grunt

#!/bin/sh

cd src
GOOS=linux GOARCH=amd64 go build -o ip-atlas.exe .
cd ..
rm -rf out
mkdir out
cp src/ip-atlas.exe out/
cp -r src/templates/ out/
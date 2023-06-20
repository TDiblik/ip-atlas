#!/bin/sh

cd src
go build -o ip-atlas.exe -ldflags="-s -w" .
cd ..
rm -rf out
mkdir out
cp src/ip-atlas.exe out/
cp -r src/templates/ out/
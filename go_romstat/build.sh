#!/usr/bin/env bash
export HmFileVersion=1.1.0
export RomStatVersion=1.1.0

echo "HmFileVersion=$HmFileVersion"
echo "RomStatVersion=$RomStatVersion"

if [ "$1" == 'win32' ]; then
    echo "amd64, windows"
    CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags="-X romstat/build.HmFileVersion=$HmFileVersion -X romstat/build.RomStatVersion=$RomStatVersion"
else
    echo "arm, linux/android"
    CGO_ENABLED=0 GOARM=7 GOARCH=arm64 GOOS=linux go build -ldflags="-X romstat/build.HmFileVersion=$HmFileVersion -X romstat/build.RomStatVersion=$RomStatVersion"
fi

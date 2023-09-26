#!/usr/bin/env bash
export HmFileVersion=1.1.0
export RomStatVersion=1.1.0

echo "HmFileVersion=$HmFileVersion"
echo "RomStatVersion=$RomStatVersion"

CGO_ENABLED=0 GOARM=7 GOARCH=arm GOOS=linux go build -ldflags="-X romstat/build.HmFileVersion=$HmFileVersion\
 -X romstat/build.RomStatVersion=$RomStatVersion"

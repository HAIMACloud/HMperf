
@echo off

set HmFileVersion=1.1.0
set RomStatVersion=1.1.0

echo HmFileVersion=%HmFileVersion%
echo RomStatVersion=%RomStatVersion%

set CGO_ENABLED=0
set GOARM=7
set GOARCH=arm
set GOOS=linux

go build -ldflags "-X romstat/build.HmFileVersion=%HmFileVersion% -X romstat/build.RomStatVersion=%RomStatVersion%"
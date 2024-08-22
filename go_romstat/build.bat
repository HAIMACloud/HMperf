
@echo off

set HmFileVersion=1.1.0
set RomStatVersion=1.2.0

echo HmFileVersion=%HmFileVersion%
echo RomStatVersion=%RomStatVersion%

set CGO_ENABLED=0
set GOARM=7
if /i "%1"=="win32" (
    echo "amd64, windows"
    set GOARCH=amd64
    set GOOS=windows
) else (
    echo "arm, linux"
    set GOARCH=arm
    set GOOS=linux
)
go build -ldflags "-X romstat/build.HmFileVersion=%HmFileVersion% -X romstat/build.RomStatVersion=%RomStatVersion%"
@echo off
setlocal
set GOPATH=%~dp0
if not exist src\github.com\VonC (
	mkdir src\github.com\VonC
)
if not exist src\github.com\VonC\seec2 (
	mklink /J src\github.com\VonC\seec2 %GOPATH%
)
pushd %~dp0
cd src\github.com\VonC\seec2
cd
go install
popd
endlocal

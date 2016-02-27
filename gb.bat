@echo off
setlocal
set GOPATH=%~dp0
if not exist src\github.com\VonC (
	mkdir src\github.com\VonC
)
if not exist src\github.com\VonC\seec (
	mklink /J src\github.com\VonC\seec %GOPATH%
)
pushd %~dp0
cd src\github.com\VonC\seec
cd
go install
popd
endlocal

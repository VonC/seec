@echo off
setlocal
set GOPATH=%~dp0/deps
set GOBIN=%~dp0/bin
go install
endlocal

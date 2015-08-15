@echo off
set go-github=src\github.com\google\go-github
if not "%GOPATH%" == "" (
	if not exist "%GOPATH%\%go-github%" (
		mklink /J "%GOPATH%\%go-github%" "%~dp0\deps\%go-github%"
		cd "%GOPATH%\%go-github%\github"
		go install
	)
)
setlocal
set src=src\github.com\VonC\godbg
if not exist "%~dp0\deps\%src%" (
	mklink /J "%~dp0\deps\%src%" "%~dp0"
)
set GOPATH=%~dp0/deps
set GOBIN=%~dp0/bin
cd "%~dp0"
go install
endlocal
doskey seec=%~dp0\bin\seec.exe $*

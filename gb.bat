@echo off
set go-github=src\github.com\google\go-github
if not "%GOPATH%" == "" (
	if not exist %GOPATH%\%go-github% (
		mklink /D %~dp0\deps\%go-github% %GOPATH%\%go-github%
		cd %GOPATH%\%go-github%
		go install
	)
)
setlocal
set GOPATH=%~dp0/deps
set GOBIN=%~dp0/bin
cd %~dp0
go install
endlocal
doskey seec=%~dp0\bin\seec.exe $*

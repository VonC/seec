@echo off
setlocal EnableDelayedExpansion
if not defined GOROOT (
        echo Environment variable GOROOT must be defined, with %%GOROOT%%\bin\go.exe
        exit /b 1
)

set PATH=C:\WINDOWS\system32;C:\WINDOWS;C:\WINDOWS\System32\Wbem
set PATH=%PATH%;%GOROOT%/bin

set prjname=%~dp0
set prjname=%prjname:~0,-1%
for %%i in ("%prjname%") do set "prjname=%%~ni"
echo prjname='%prjname%'

mkdir /F %~dp0..\_gopaths\%prjname%_gopath\src 2> NUL
pushd %~dp0..\_gopaths\%prjname%_gopath
set GOPATH=%CD%
popd

if not exist %GOPATH%\src\%prjname% (
	mklink /J %GOPATH%\src\%prjname% %~dp0
)
if not exist bin (
	mklink /J bin %GOPATH%\bin
)
if not exist pkg (
	mklink /J pkg %GOPATH%\pkg
)
pushd %GOPATH%\src\%prjname%
cd
go install
popd
endlocal

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

mkdir /F %~dp0..\_gopaths\%prjname%_gopath 2> NUL
pushd %~dp0..\_gopaths\%prjname%_gopath
set GOPATH=%CD%
popd

if not exist bin (
	mklink /J bin %GOPATH%\bin
)
if not exist bin\%prjname%.exe (
	echo "prjname='%prjname%' does not exist yet (not compiled)"
	exit /b 1
)
cmd /v /c "set dbg=1 && bin\%prjname%.exe a58a8e3f7156dadd5d5e9643168545ff057c111a"
endlocal
